// ================================================================
// PACOTE SERVICES - ONBOARDING SERVICE (CADASTRO PÚBLICO DO SITE)
// ================================================================
// Atende o POST /api/public/associations, chamado pelo site-cannacare.
//
// FLUXO:
//   1. Revalida o CNPJ no servidor (nunca confia só no front)
//   2. Confere que CNPJ e email ainda não estão cadastrados
//   3. Cria a associação:
//        - plano gratuito -> status 'active' na hora
//        - plano premium  -> status 'pending' até o pagamento confirmar
//   4. Cria o usuário admin (SEM senha utilizável ainda) + um token
//      de convite (link de primeiro acesso)
//   5. Plano gratuito -> manda o email de convite JÁ
//      Plano premium  -> cria a preferência no Mercado Pago e devolve
//                         a URL de checkout (o email só sai depois,
//                         quando o webhook confirmar o pagamento)
// ================================================================

package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"cannacare-backend/internal/models"
	"cannacare-backend/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OnboardingService struct {
	db             *gorm.DB
	emailService   *EmailService
	paymentService *PaymentService
}

func NewOnboardingService(db *gorm.DB, emailService *EmailService, paymentService *PaymentService) *OnboardingService {
	return &OnboardingService{
		db:             db,
		emailService:   emailService,
		paymentService: paymentService,
	}
}

// ================================================================
// STRUCTS
// ================================================================

type EnderecoRequest struct {
	Logradouro string `json:"logradouro"`
	Numero     string `json:"numero"`
	Bairro     string `json:"bairro"`
	Municipio  string `json:"municipio"`
	UF         string `json:"uf"`
	CEP        string `json:"cep"`
}

// OnboardingRequest - payload que vem do formulário do site
type OnboardingRequest struct {
	Plano              string          `json:"plano" validate:"required,oneof=gratuito premium"`
	CNPJ               string          `json:"cnpj" validate:"required"`
	RazaoSocial        string          `json:"razaoSocial"`
	NomeFantasia       string          `json:"nomeFantasia"`
	ResponsavelNome    string          `json:"responsavelNome" validate:"required,min=3"`
	ResponsavelEmail   string          `json:"responsavelEmail" validate:"required,email"`
	ResponsavelTelefone string         `json:"responsavelTelefone" validate:"required"`
	Endereco           EnderecoRequest `json:"endereco"`
}

type OnboardingResponseData struct {
	AssociationID string `json:"association_id,omitempty"`
	CheckoutURL   string `json:"checkout_url,omitempty"`
}

// Preços em reais - plano premium mensal (mesma regra usada na página de planos)
const precoPremiumMensal = 1.00

// ================================================================
// CREATE
// ================================================================
func (s *OnboardingService) Create(req OnboardingRequest) (*OnboardingResponseData, error) {
	// --- 1. Revalidar CNPJ no servidor (nunca confiar no front) ---
	dadosCNPJ, err := utils.ConsultarCNPJ(req.CNPJ)
	if err != nil {
		return nil, err
	}

	// --- 2. CNPJ já cadastrado? ---
	var existingAssoc models.Association
	if err := s.db.Where("cnpj = ?", dadosCNPJ.CNPJ).First(&existingAssoc).Error; err == nil {
		return nil, errors.New("já existe uma associação cadastrada com este CNPJ")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- 3. Email já cadastrado? (login busca globalmente por email) ---
	if !utils.IsValidEmail(req.ResponsavelEmail) {
		return nil, errors.New("email do responsável inválido")
	}
	var existingUser models.User
	if err := s.db.Where("email = ?", req.ResponsavelEmail).First(&existingUser).Error; err == nil {
		return nil, errors.New("este email já está cadastrado em outra associação")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	isPremium := req.Plano == "premium"

	nome := dadosCNPJ.NomeFantasia
	if nome == "" {
		nome = dadosCNPJ.RazaoSocial
	}

	// --- 4. Criar a associação ---
	association := &models.Association{
		Name:    nome,
		CNPJ:    dadosCNPJ.CNPJ,
		Email:   req.ResponsavelEmail,
		Phone:   req.ResponsavelTelefone,
		Address: dadosCNPJ.Endereco,
		Plan:    "basic", // só vira "premium" de fato quando o pagamento confirmar
	}
	if isPremium {
		association.Status = "pending"
		association.PatientLimit = 50
		association.UserLimit = 3
	} else {
		association.Status = "active"
		association.PatientLimit = 50
		association.UserLimit = 3
	}

	if err := s.db.Create(association).Error; err != nil {
		return nil, fmt.Errorf("erro ao criar associação: %w", err)
	}

	// --- 5. Criar o usuário admin (senha ainda não utilizável) ---
	// Gera um hash de uma senha aleatória que ninguém sabe — o usuário
	// só ganha acesso de verdade quando definir a própria senha pelo
	// link de convite.
	randomPlaceholder := uuid.New().String() + uuid.New().String()
	placeholderHash, err := bcrypt.GenerateFromPassword([]byte(randomPlaceholder), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("erro interno ao preparar usuário")
	}

	adminUser := &models.User{
		AssociationID: association.ID,
		Name:          req.ResponsavelNome,
		Email:         req.ResponsavelEmail,
		PasswordHash:  string(placeholderHash),
		Role:          "admin",
		IsActive:      !isPremium, // premium só fica ativo quando o pagamento confirmar
	}
	if err := s.db.Create(adminUser).Error; err != nil {
		return nil, fmt.Errorf("erro ao criar usuário admin: %w", err)
	}

	// --- 6. Criar o token de convite (link de primeiro acesso) ---
	rawToken, tokenHash, err := generateInviteToken()
	if err != nil {
		return nil, errors.New("erro interno ao gerar convite")
	}
	invite := &models.InviteToken{
		UserID:    adminUser.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.db.Create(invite).Error; err != nil {
		return nil, fmt.Errorf("erro ao gerar convite: %w", err)
	}

	inviteLink := fmt.Sprintf("%s/primeiro-acesso?token=%s", os.Getenv("FRONTEND_APP_URL"), rawToken)

	// --- 7. Plano gratuito: ativa e manda o email agora ---
	if !isPremium {
		if err := s.emailService.SendInviteEmail(req.ResponsavelEmail, req.ResponsavelNome, nome, inviteLink, "gratuito"); err != nil {
			// Não falha o cadastro por causa do email — a associação já existe.
			// Fica registrado pra investigar/reenviar depois.
			fmt.Printf("⚠️ erro ao enviar email de convite para %s: %v\n", req.ResponsavelEmail, err)
		}
		return &OnboardingResponseData{AssociationID: association.ID.String()}, nil
	}

	// --- 8. Plano premium: gera o checkout do Mercado Pago ---
	checkoutURL, preferenceID, err := s.paymentService.CreateCheckoutPreference(association.ID, req.ResponsavelEmail, precoPremiumMensal)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar pagamento: %w", err)
	}

	association.PaymentReference = preferenceID
	s.db.Save(association)

	// O email do plano premium só sai quando o webhook confirmar o
	// pagamento (ver EmailService.SendPlanActivatedEmail) — não aqui.

	return &OnboardingResponseData{
		AssociationID: association.ID.String(),
		CheckoutURL:   checkoutURL,
	}, nil
}

// ================================================================
// INVITE INFO - dados mostrados na tela de primeiro acesso
// ================================================================
type InviteInfo struct {
	UserName        string `json:"user_name"`
	Email           string `json:"email"`
	AssociationName string `json:"association_name"`
}

// ================================================================
// VALIDATEINVITETOKEN - GET /api/public/invite/{token}
// ================================================================
// Confere se o token é válido (existe, não expirou, não foi usado)
// e devolve os dados pra tela mostrar "Olá, Fulano da Associação X".
func (s *OnboardingService) ValidateInviteToken(rawToken string) (*InviteInfo, error) {
	invite, err := s.findValidInvite(rawToken)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.db.Where("id = ?", invite.UserID).First(&user).Error; err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	var association models.Association
	if err := s.db.Where("id = ?", user.AssociationID).First(&association).Error; err != nil {
		return nil, errors.New("associação não encontrada")
	}

	return &InviteInfo{
		UserName:        user.Name,
		Email:           user.Email,
		AssociationName: association.Name,
	}, nil
}

// ================================================================
// REDEEMINVITETOKEN - POST /api/public/invite/{token}
// ================================================================
// Define a senha de verdade do usuário e marca o token como usado
// (uso único — não dá pra reaproveitar o mesmo link depois).
func (s *OnboardingService) RedeemInviteToken(rawToken, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("a senha precisa ter pelo menos 6 caracteres")
	}

	invite, err := s.findValidInvite(rawToken)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("erro ao criptografar senha")
	}

	var user models.User
	if err := s.db.Where("id = ?", invite.UserID).First(&user).Error; err != nil {
		return errors.New("usuário não encontrado")
	}
	user.PasswordHash = string(hashedPassword)
	user.IsActive = true
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	now := time.Now()
	invite.UsedAt = &now
	return s.db.Save(invite).Error
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// generateInviteToken gera um token aleatório seguro. Devolve o valor
// PURO (vai no link do email) e o HASH (o que fica salvo no banco).
func generateInviteToken() (rawToken string, tokenHash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	rawToken = hex.EncodeToString(bytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(hash[:])

	return rawToken, tokenHash, nil
}

// findValidInvite busca o token pelo hash e confere expiração/uso único.
func (s *OnboardingService) findValidInvite(rawToken string) (*models.InviteToken, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	var invite models.InviteToken
	if err := s.db.Where("token_hash = ?", tokenHash).First(&invite).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("link inválido")
		}
		return nil, err
	}

	if invite.UsedAt != nil {
		return nil, errors.New("este link já foi utilizado")
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, errors.New("este link expirou, peça um novo acesso")
	}

	return &invite, nil
}

// ================================================================
// PROCESSPAYMENTWEBHOOK - ponto de entrada único para o webhook
// ================================================================
// Busca o pagamento DE VERDADE na API do Mercado Pago (nunca confia
// só no que o webhook mandou) e, se estiver aprovado, ativa o plano.
func (s *OnboardingService) ProcessPaymentWebhook(mpPaymentID string) error {
	payment, err := s.paymentService.GetPayment(mpPaymentID)
	if err != nil {
		return err
	}

	// Só nos importa quando o pagamento foi de fato aprovado.
	// Outros status (pending, rejected, in_process...) são ignorados
	// aqui — o Mercado Pago vai mandar um novo webhook se/quando
	// o status mudar para approved.
	if payment.Status != "approved" {
		return nil
	}

	if payment.ExternalReference == "" {
		return errors.New("pagamento aprovado sem external_reference")
	}

	associationID, err := uuid.Parse(payment.ExternalReference)
	if err != nil {
		return fmt.Errorf("external_reference inválido: %w", err)
	}

	return s.ActivatePremiumPlan(associationID, payment.ID)
}

// ================================================================
// ACTIVATEPREMIUMPLAN - chamado pelo webhook do Mercado Pago
// ================================================================
// Ativa o plano premium de uma associação depois que o pagamento foi
// CONFIRMADO de verdade na API do Mercado Pago (não confia no webhook
// sozinho, isso já foi checado antes de chamar essa função).
//
// É idempotente: se a associação já estiver ativa, não faz nada de
// novo (o Mercado Pago pode reenviar o mesmo webhook várias vezes).
func (s *OnboardingService) ActivatePremiumPlan(associationID uuid.UUID, paymentReference string) error {
	var association models.Association
	if err := s.db.Where("id = ?", associationID).First(&association).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("associação não encontrada")
		}
		return err
	}

	// --- Idempotência: ESSE pagamento específico já foi processado ---
	// (não checamos status/plan aqui, porque numa RENOVAÇÃO a associação
	// já está ativa/premium mesmo assim — o que não pode repetir é
	// processar o MESMO pagamento duas vezes, caso o Mercado Pago
	// reenvie a notificação).
	if paymentReference != "" && association.PaymentReference == paymentReference {
		return nil
	}

	// --- Renovação x primeira ativação ---
	// Se já era premium/ativa antes desse pagamento, é renovação: soma
	// 12 meses ao prazo que ainda resta (em vez de zerar), e não manda
	// convite de novo (o admin já tem senha configurada).
	isRenewal := association.Status == "active" && association.Plan == "premium"

	now := time.Now()
	baseTime := now
	if isRenewal && association.PlanExpiresAt != nil && association.PlanExpiresAt.After(now) {
		baseTime = *association.PlanExpiresAt
	}
	expiresAt := baseTime.AddDate(1, 0, 0) // 12 meses

	association.Plan = "premium"
	association.Status = "active"
	association.PatientLimit = 999999
	association.UserLimit = 10
	if !isRenewal {
		association.PlanActivatedAt = &now
	}
	association.PlanExpiresAt = &expiresAt
	association.PaymentReference = paymentReference

	if err := s.db.Save(&association).Error; err != nil {
		return err
	}

	// --- Buscar o admin da associação ---
	var adminUser models.User
	if err := s.db.Where("association_id = ? AND role = ?", association.ID, "admin").First(&adminUser).Error; err != nil {
		return fmt.Errorf("associação ativada, mas admin não encontrado: %w", err)
	}

	// --- Renovação: só confirma por email, sem link de convite novo ---
	if isRenewal {
		if err := s.emailService.SendPlanRenewedEmail(adminUser.Email, adminUser.Name, association.Name, expiresAt); err != nil {
			fmt.Printf("⚠️ erro ao enviar email de renovação para %s: %v\n", adminUser.Email, err)
		}
		return nil
	}

	// --- Primeira ativação: gerar link de convite (o do cadastro nunca foi enviado) ---
	rawToken, tokenHash, err := generateInviteToken()
	if err != nil {
		return fmt.Errorf("associação ativada, mas erro ao gerar convite: %w", err)
	}
	invite := &models.InviteToken{
		UserID:    adminUser.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.db.Create(invite).Error; err != nil {
		return fmt.Errorf("associação ativada, mas erro ao salvar convite: %w", err)
	}

	inviteLink := fmt.Sprintf("%s/primeiro-acesso?token=%s", os.Getenv("FRONTEND_APP_URL"), rawToken)

	// --- Enviar o email de ativação ---
	if err := s.emailService.SendPlanActivatedEmail(adminUser.Email, adminUser.Name, association.Name, inviteLink); err != nil {
		// Não falha a ativação por causa do email — fica registrado pra reenviar depois.
		fmt.Printf("⚠️ erro ao enviar email de ativação para %s: %v\n", adminUser.Email, err)
	}

	return nil
}

// ================================================================
// CREATERENEWALCHECKOUT - endpoint autenticado (cliente já logado)
// ================================================================
// Diferente do onboarding: aqui a associação JÁ EXISTE e está
// logada — não precisa de CNPJ, formulário nem nada disso. Só gera
// um novo checkout do Mercado Pago pra renovar por mais 12 meses.
// O webhook, ao confirmar, chama ActivatePremiumPlan de novo, que já
// sabe estender o prazo em vez de zerar (ver isRenewal ali em cima).
func (s *OnboardingService) CreateRenewalCheckout(associationID uuid.UUID) (checkoutURL string, err error) {
	var association models.Association
	if err := s.db.Where("id = ?", associationID).First(&association).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("associação não encontrada")
		}
		return "", err
	}

	checkoutURL, preferenceID, err := s.paymentService.CreateCheckoutPreference(association.ID, association.Email, precoPremiumMensal)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar pagamento: %w", err)
	}

	// Guarda a preferência gerada (não o pagamento em si ainda — isso só
	// muda quando o webhook confirmar de verdade, com o payment_id real).
	_ = preferenceID

	return checkoutURL, nil
}