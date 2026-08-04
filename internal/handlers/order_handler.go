// ================================================================
// PACOTE HANDLERS - ORDER HANDLER
// ================================================================
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderService *services.OrderService
	validator    *validator.Validate
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		validator:    validator.New(),
	}
}

func (h *OrderHandler) extractAssociationID(r *http.Request) (uuid.UUID, error) {
	associationIDStr, _ := r.Context().Value(middleware.AssociationIDKey).(string)
	if associationIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(associationIDStr)
}

// POST /api/orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	var req services.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID do usuário")
		return
	}

	order, err := h.orderService.Create(associationID, req, userID)
	if err != nil {
		log.Printf("❌ Erro ao criar pedido: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, order)
}

// GET /api/orders/{id}
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	order, err := h.orderService.GetByID(associationID, id)
	if err != nil {
		if err.Error() == "pedido não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar pedido: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar pedido")
		return
	}

	utils.SendSuccess(w, http.StatusOK, order)
}

// GET /api/orders
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	req := services.ListOrderRequest{
		PatientID:      r.URL.Query().Get("patient_id"),
		PrescriptionID: r.URL.Query().Get("prescription_id"),
		Status:         r.URL.Query().Get("status"),
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil {
			req.Page = page
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil {
			req.Limit = limit
		}
	}

	orders, total, err := h.orderService.List(associationID, req)
	if err != nil {
		log.Printf("❌ Erro ao listar pedidos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar pedidos")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": orders,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// GET /api/orders/patient/{id}
func (h *OrderHandler) GetByPatient(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	orders, err := h.orderService.GetByPatient(associationID, patientID)
	if err != nil {
		log.Printf("❌ Erro ao buscar pedidos do paciente: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar pedidos do paciente")
		return
	}

	utils.SendSuccess(w, http.StatusOK, orders)
}

// PATCH /api/orders/{id}/status
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req services.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID do usuário")
		return
	}

	order, err := h.orderService.UpdateStatus(associationID, id, req, userID)
	if err != nil {
		if err.Error() == "pedido não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar status do pedido: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, order)
}

// PATCH /api/orders/{id}/tracking
func (h *OrderHandler) UpdateTracking(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req services.UpdateTrackingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	order, err := h.orderService.UpdateTracking(associationID, id, req)
	if err != nil {
		if err.Error() == "pedido não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar rastreio: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, order)
}

// POST /api/orders/{id}/label
func (h *OrderHandler) GenerateLabel(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	labelURL, err := h.orderService.GenerateLabel(associationID, id)
	if err != nil {
		if err.Error() == "pedido não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao gerar etiqueta: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"label_url": labelURL,
		"message":   "Etiqueta gerada com sucesso",
	})
}