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
	"log"
	"net/http"

	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type PublicHandler struct {
	onboardingService *services.OnboardingService
	validator         *validator.Validate
}

func NewPublicHandler(onboardingService *services.OnboardingService) *PublicHandler {
	return &PublicHandler{
		onboardingService: onboardingService,
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
// POST /api/public/billing/webhook/mercadopago
// ================================================================
// O Mercado Pago manda essa notificação de duas formas possíveis:
//   1. Query params (formato antigo/IPN): ?topic=payment&id=123
//   2. Corpo JSON (formato novo): {"type":"payment","data":{"id":"123"}}
//
// ⚠️ IMPORTANTE: sempre respondemos 200 rapidinho, mesmo se o
// processamento falhar internamente — se responder erro, o Mercado
// Pago fica reenviando a mesma notificação sem parar. Erros reais
// ficam só no log do servidor pra investigar depois.
type mercadoPagoWebhookBody struct {
	Type string `json:"type"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (h *PublicHandler) MercadoPagoWebhook(w http.ResponseWriter, r *http.Request) {
	var paymentID string
	var eventType string

	// --- 1. Tentar ler do corpo JSON (formato novo) ---
	var body mercadoPagoWebhookBody
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.Data.ID != "" {
		paymentID = body.Data.ID
		eventType = body.Type
	}

	// --- 2. Se não veio no corpo, tentar query params (formato antigo) ---
	if paymentID == "" {
		paymentID = r.URL.Query().Get("id")
		eventType = r.URL.Query().Get("topic")
		if eventType == "" {
			eventType = r.URL.Query().Get("type")
		}
	}

	// --- 3. Só nos interessa notificação de pagamento ---
	if eventType != "payment" || paymentID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// --- 4. Processar (busca o pagamento de verdade e ativa se aprovado) ---
	if err := h.onboardingService.ProcessPaymentWebhook(paymentID); err != nil {
		log.Printf("❌ erro ao processar webhook do Mercado Pago (payment_id=%s): %v", paymentID, err)
		// Ainda assim respondemos 200 — o erro já está logado, e devolver
		// erro só faria o Mercado Pago reenviar a mesma notificação à toa.
	}

	w.WriteHeader(http.StatusOK)
}