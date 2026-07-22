// ================================================================
// PACOTE HANDLERS - FINANCIAL HANDLER
// ================================================================
// Camada HTTP que lida com as requisições financeiras.
//
// ENDPOINTS:
//   POST   /api/financial/subscriptions           - Criar anuidade
//   GET    /api/financial/subscriptions           - Listar anuidades
//   GET    /api/financial/subscriptions/{id}      - Buscar anuidade
//   POST   /api/financial/payments                - Registrar pagamento
//   GET    /api/financial/payments                - Listar pagamentos
//   GET    /api/financial/payments/{id}           - Buscar pagamento
//   PATCH  /api/financial/payments/{id}/status    - Atualizar status
//   GET    /api/financial/patient/{id}            - Status financeiro
//   GET    /api/financial/overdue                 - Anuidades em atraso
// ================================================================

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type FinancialHandler struct {
	financialService *services.FinancialService
	validator        *validator.Validate
}

func NewFinancialHandler(financialService *services.FinancialService) *FinancialHandler {
	return &FinancialHandler{
		financialService: financialService,
		validator:        validator.New(),
	}
}

// ================================================================
// SUBSCRIPTION HANDLERS
// ================================================================

// CreateSubscription - POST /api/financial/subscriptions
func (h *FinancialHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req services.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	subscription, err := h.financialService.CreateSubscription(req)
	if err != nil {
		log.Printf("❌ Erro ao criar anuidade: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, subscription)
}

// ListSubscriptions - GET /api/financial/subscriptions
func (h *FinancialHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	req := services.ListSubscriptionRequest{
		PatientID: r.URL.Query().Get("patient_id"),
		Status:    r.URL.Query().Get("status"),
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, _ := strconv.Atoi(pageStr)
		req.Page = page
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		req.Limit = limit
	}

	subscriptions, total, err := h.financialService.ListSubscriptions(req)
	if err != nil {
		log.Printf("❌ Erro ao listar anuidades: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar anuidades")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": subscriptions,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// GetSubscriptionByID - GET /api/financial/subscriptions/{id}
func (h *FinancialHandler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	subscription, err := h.financialService.GetSubscriptionByID(id)
	if err != nil {
		if err.Error() == "anuidade não encontrada" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar anuidade: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar anuidade")
		return
	}

	utils.SendSuccess(w, http.StatusOK, subscription)
}

// ================================================================
// PAYMENT HANDLERS
// ================================================================

// CreatePayment - POST /api/financial/payments
func (h *FinancialHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req services.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	payment, err := h.financialService.CreatePayment(req)
	if err != nil {
		log.Printf("❌ Erro ao criar pagamento: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, payment)
}

// ListPayments - GET /api/financial/payments
func (h *FinancialHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	req := services.ListPaymentRequest{
		PatientID:   r.URL.Query().Get("patient_id"),
		PaymentType: r.URL.Query().Get("payment_type"),
		Status:      r.URL.Query().Get("status"),
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, _ := strconv.Atoi(pageStr)
		req.Page = page
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		req.Limit = limit
	}

	payments, total, err := h.financialService.ListPayments(req)
	if err != nil {
		log.Printf("❌ Erro ao listar pagamentos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar pagamentos")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": payments,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// GetPaymentByID - GET /api/financial/payments/{id}
func (h *FinancialHandler) GetPaymentByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	payment, err := h.financialService.GetPaymentByID(id)
	if err != nil {
		if err.Error() == "pagamento não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar pagamento: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar pagamento")
		return
	}

	utils.SendSuccess(w, http.StatusOK, payment)
}

// UpdatePaymentStatus - PATCH /api/financial/payments/{id}/status
func (h *FinancialHandler) UpdatePaymentStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req services.UpdatePaymentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	payment, err := h.financialService.UpdatePaymentStatus(id, req)
	if err != nil {
		if err.Error() == "pagamento não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar status do pagamento: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, payment)
}

// ================================================================
// DASHBOARD & STATUS
// ================================================================

// GetPatientFinancialStatus - GET /api/financial/patient/{id}
func (h *FinancialHandler) GetPatientFinancialStatus(w http.ResponseWriter, r *http.Request) {
	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	status, err := h.financialService.GetPatientFinancialStatus(patientID)
	if err != nil {
		if err.Error() == "paciente não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar status financeiro: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar status financeiro")
		return
	}

	utils.SendSuccess(w, http.StatusOK, status)
}

// GetOverdueSubscriptions - GET /api/financial/overdue
func (h *FinancialHandler) GetOverdueSubscriptions(w http.ResponseWriter, r *http.Request) {
	overdue, err := h.financialService.GetOverdueSubscriptions()
	if err != nil {
		log.Printf("❌ Erro ao buscar anuidades em atraso: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar anuidades em atraso")
		return
	}

	utils.SendSuccess(w, http.StatusOK, overdue)
}
