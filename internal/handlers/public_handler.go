// ================================================================
// PACOTE HANDLERS - PUBLIC HANDLER (SEM AUTENTICAÇÃO)
// ================================================================
// Endpoints chamados pelo site-cannacare, ANTES de existir login.
//
//   POST /api/public/associations - cadastro da associação (site)
//
// ⚠️ Fica FORA do grupo de rotas que usa middleware.AuthMiddleware
// no main.go — não tem token JWT nessa etapa.
// ================================================================

package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type PublicHandler struct {
	onboardingService *services.OnboardingService
	stripeService     *services.StripeService
	validator         *validator.Validate
}

func NewPublicHandler(onboardingService *services.OnboardingService, stripeService *services.StripeService) *PublicHandler {
	return &PublicHandler{
		onboardingService: onboardingService,
		stripeService:     stripeService,
		validator:         validator.New(),
	}
}

// ================================================================
// POST /api/public/associations
// ================================================================
func (h *PublicHandler) CreateAssociation(w http.ResponseWriter, r *http.Request) {
	var req services.OnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	data, err := h.onboardingService.Create(req)
	if err != nil {
		log.Printf("❌ erro no cadastro público: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, data)
}

// ================================================================
// GET /api/public/invite/{token}
// ================================================================
// Chamado quando a tela de primeiro acesso carrega, pra mostrar
// "Olá, Fulano da Associação X" antes de pedir a senha.
func (h *PublicHandler) ValidateInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	info, err := h.onboardingService.ValidateInviteToken(token)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, info)
}

type redeemInviteRequest struct {
	Password string `json:"password" validate:"required,min=6"`
}

// ================================================================
// POST /api/public/invite/{token}
// ================================================================
// Define a senha de verdade. Depois disso o token não serve mais.
func (h *PublicHandler) RedeemInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var req redeemInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	if err := h.onboardingService.RedeemInviteToken(token, req.Password); err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Senha definida com sucesso",
	})
}

// ================================================================
// POST /api/public/billing/webhook/stripe
// ================================================================
// O Stripe assina cada webhook com HMAC (header "Stripe-Signature"),
// usando a STRIPE_WEBHOOK_SECRET configurada — é assim que sabemos
// que a notificação é de verdade, e não alguém forjando um POST.
//
// ⚠️ IMPORTANTE: sempre respondemos 200 rapidinho, mesmo se o
// processamento falhar internamente — se responder erro, o Stripe
// fica reenviando a mesma notificação sem parar. Erros reais ficam
// só no log do servidor pra investigar depois.
func (h *PublicHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	// --- 1. Lê o corpo BRUTO (a assinatura é calculada em cima dos
	// bytes exatos que o Stripe mandou — não pode re-serializar) ---
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ erro ao ler corpo do webhook do Stripe: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// --- 2. Verifica a assinatura ---
	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := h.stripeService.VerifyWebhookSignature(payload, sigHeader)
	if err != nil {
		log.Printf("❌ webhook do Stripe com assinatura inválida: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// --- 3. Só nos interessa quando o checkout foi concluído E pago ---
	if event.Type != "checkout.session.completed" || event.Data.Object.PaymentStatus != "paid" {
		w.WriteHeader(http.StatusOK)
		return
	}

	associationIDStr := event.Data.Object.Metadata.AssociationID
	if associationIDStr == "" {
		log.Printf("❌ webhook do Stripe aprovado sem association_id no metadata")
		w.WriteHeader(http.StatusOK)
		return
	}

	associationID, err := uuid.Parse(associationIDStr)
	if err != nil {
		log.Printf("❌ association_id inválido no webhook do Stripe: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// --- 4. Ativa o plano ---
	if err := h.onboardingService.ActivatePremiumPlan(associationID, event.Data.Object.ID); err != nil {
		log.Printf("❌ erro ao ativar plano (association_id=%s): %v", associationID, err)
		// Ainda assim respondemos 200 — o erro já está logado, e devolver
		// erro só faria o Stripe reenviar a mesma notificação à toa.
	}

	w.WriteHeader(http.StatusOK)
}