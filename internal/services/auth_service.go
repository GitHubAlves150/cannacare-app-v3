// ================================================================
// CANNACARE - AUTH SERVICE
// ================================================================
// Camada de serviço responsável pela autenticação e registro de usuários.
//
// RESPONSABILIDADES:
//   1. Registrar novos usuários (com associação)
//   2. Autenticar usuários (login)
//   3. Gerar tokens JWT
//   4. Gerenciar sessões (last_login_at)
//
// FLUXO DE REGISTRO (Onboarding):
//   1. A associação se cadastra (nome, CNPJ, email, senha)
//   2. O sistema cria a ASSOCIATION na tabela associations
//   3. O sistema cria o USUÁRIO ADMIN vinculado à associação
//   4. O usuário pode fazer login e começar a usar
//
// FLUXO DE LOGIN:
//   1. Usuário envia email + senha
//   2. Busca o usuário no banco
//   3. Verifica a senha (bcrypt)
//   4. Gera o JWT com association_id
//   5. Retorna o token para o cliente
// ================================================================

package services

import (
	"errors"
	"time"

	"cannacare-backend/internal/models"
	"cannacare-backend/internal/utils"
	"cannacare-backend/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT AUTHSERVICE
// ================================================================
type AuthService struct {
	db         *gorm.DB
	jwtService *jwt.JWTService
}

// ================================================================
// FUNÇÃO NEWAUTHSERVICE()
// ================================================================
func NewAuthService(db *gorm.DB, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		db:         db,
		jwtService: jwtService,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

// RegisterRequest - Dados para registrar uma nova associação + admin
type RegisterRequest struct {
	// Dados da associação
	AssociationName string `json:"association_name" validate:"required,min=3"`
	CNPJ            string `json:"cnpj" validate:"required"`
	Phone           string `json:"phone" validate:"omitempty"`

	// Dados do usuário admin
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest - Dados para fazer login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse - Resposta do login (token + dados do usuário)
type AuthResponse struct {
	Token     string        `json:"token"`
	TokenType string        `json:"token_type"`
	ExpiresIn int64         `json:"expires_in"`
	User      *UserResponse `json:"user"`
}

// UserResponse - Dados do usuário (sem senha)
type UserResponse struct {
	ID            string     `json:"id"`
	AssociationID string     `json:"association_id"` // ← MULTI-TENANCY!
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Role          string     `json:"role"`
	IsActive      bool       `json:"is_active"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ================================================================
// FUNÇÃO REGISTER() - ONBOARDING DE NOVAS ASSOCIAÇÕES
// ================================================================
// Registra uma nova associação e cria o usuário admin.
//
// FLUXO COMPLETO:
//  1. Valida os dados (email, CNPJ, senha)
//  2. Verifica se o email/CNPJ já existe
//  3. Cria a associação na tabela associations
//  4. Cria o usuário admin vinculado à associação
//  5. Retorna os dados do usuário criado
//
// ⚠️ IMPORTANTE: O association_id do usuário é o ID da associação criada!
func (s *AuthService) Register(req RegisterRequest) (*UserResponse, error) {
	// --- PASSO 1: Validar email ---
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("email inválido")
	}

	// --- PASSO 2: Verificar se o email já existe ---
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email já cadastrado")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- PASSO 3: Verificar se o CNPJ já existe ---
	var existingAssociation models.Association
	if err := s.db.Where("cnpj = ?", req.CNPJ).First(&existingAssociation).Error; err == nil {
		return nil, errors.New("CNPJ já cadastrado")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- PASSO 4: Criptografar a senha ---
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("erro ao criptografar senha")
	}

	// --- PASSO 5: Criar a associação (SEM PONTEIRO) ---
	association := models.Association{ // ← SEM & (sem ponteiro)
		Name:         req.AssociationName,
		CNPJ:         req.CNPJ,
		Email:        req.Email,
		Phone:        req.Phone,
		Plan:         "basic",
		Status:       "pending",
		PatientLimit: 50,
		UserLimit:    3, // ← ADICIONAR LIMITE DE USUÁRIOS
	}

	if err := s.db.Create(&association).Error; err != nil { // ← PASSAR PONTEIRO PARA O CREATE
		return nil, err
	}

	// --- PASSO 6: Criar o usuário admin ---
	user := &models.User{
		AssociationID: association.ID,
		Name:          req.Name,
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		Role:          "admin",
		IsActive:      true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	// --- PASSO 7: Retornar os dados do usuário ---
	return &UserResponse{
		ID:            user.ID.String(),
		AssociationID: user.AssociationID.String(),
		Name:          user.Name,
		Email:         user.Email,
		Role:          user.Role,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
	}, nil
}

// ================================================================
// FUNÇÃO LOGIN()
// ================================================================
// Autentica um usuário e gera um token JWT.
//
// FLUXO COMPLETO:
//  1. Valida o email
//  2. Busca o usuário pelo email
//  3. Verifica se o usuário está ativo
//  4. Compara a senha (bcrypt)
//  5. Gera o token JWT com association_id
//  6. Atualiza o last_login_at
//  7. Retorna o token + dados do usuário
//
// ⚠️ IMPORTANTE: O association_id é extraído do usuário e colocado no JWT!
// Isso garante que todas as requisições futuras saibam qual associação
// os dados pertencem.
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	// --- PASSO 1: Validar email ---
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("email inválido")
	}

	// --- PASSO 2: Buscar usuário pelo email ---
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("email ou senha incorretos")
		}
		return nil, err
	}

	// --- PASSO 3: Verificar se o usuário está ativo ---
	if !user.IsActive {
		return nil, errors.New("usuário desativado, contate o administrador")
	}

	// --- PASSO 4: Comparar a senha com o hash ---
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("email ou senha incorretos")
	}

	// --- PASSO 5: Gerar token JWT com association_id ---
	// ⚠️ O association_id é ESSENCIAL para o multi-tenancy!
	// Ele será usado em todas as requisições para filtrar dados.
	token, err := s.jwtService.GenerateToken(
		user.ID,
		user.AssociationID, // ← MULTI-TENANCY: association_id vai no JWT!
		user.Email,
		user.Name,
		user.Role,
	)
	if err != nil {
		return nil, errors.New("erro ao gerar token de autenticação")
	}

	// --- PASSO 6: Atualizar LastLoginAt ---
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	// --- PASSO 7: Calcular tempo de expiração ---
	expiresInSeconds := int64(s.jwtService.ExpiresIn().Seconds())

	return &AuthResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: expiresInSeconds,
		User: &UserResponse{
			ID:            user.ID.String(),
			AssociationID: user.AssociationID.String(), // ← MULTI-TENANCY!
			Name:          user.Name,
			Email:         user.Email,
			Role:          user.Role,
			IsActive:      user.IsActive,
			LastLoginAt:   user.LastLoginAt,
			CreatedAt:     user.CreatedAt,
		},
	}, nil
}
