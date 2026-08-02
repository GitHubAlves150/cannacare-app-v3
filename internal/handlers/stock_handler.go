// ================================================================
// PACOTE HANDLERS - STOCK HANDLER
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

type StockHandler struct {
	stockService *services.StockService
	validator    *validator.Validate
}

func NewStockHandler(stockService *services.StockService) *StockHandler {
	return &StockHandler{
		stockService: stockService,
		validator:    validator.New(),
	}
}

func (h *StockHandler) extractAssociationID(r *http.Request) (uuid.UUID, error) {
	associationIDStr, _ := r.Context().Value(middleware.AssociationIDKey).(string)
	if associationIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(associationIDStr)
}

// POST /api/stock/lots
func (h *StockHandler) CreateLot(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	var req services.CreateLotRequest
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

	lot, err := h.stockService.CreateLot(associationID, req, userID)
	if err != nil {
		log.Printf("❌ Erro ao criar lote: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, lot)
}

// GET /api/stock/lots/{id}
func (h *StockHandler) GetLotByID(w http.ResponseWriter, r *http.Request) {
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

	lot, err := h.stockService.GetLotByID(associationID, id)
	if err != nil {
		if err.Error() == "lote não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar lote: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar lote")
		return
	}

	utils.SendSuccess(w, http.StatusOK, lot)
}

// GET /api/stock/lots
func (h *StockHandler) ListLots(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	req := services.ListLotRequest{
		ProductID: r.URL.Query().Get("product_id"),
	}

	if isExpiredStr := r.URL.Query().Get("is_expired"); isExpiredStr != "" {
		isExpired, err := strconv.ParseBool(isExpiredStr)
		if err == nil {
			req.IsExpired = &isExpired
		}
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

	lots, total, err := h.stockService.ListLots(associationID, req)
	if err != nil {
		log.Printf("❌ Erro ao listar lotes: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar lotes")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": lots,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// POST /api/stock/adjust
func (h *StockHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	var req services.StockAdjustRequest
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

	movement, err := h.stockService.AdjustStock(associationID, req, userID)
	if err != nil {
		log.Printf("❌ Erro ao ajustar estoque: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, movement)
}

// GET /api/stock/movements
func (h *StockHandler) GetMovements(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	req := services.ListMovementRequest{
		ProductLotID: r.URL.Query().Get("product_lot_id"),
		Type:         r.URL.Query().Get("type"),
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

	movements, total, err := h.stockService.GetMovements(associationID, req)
	if err != nil {
		log.Printf("❌ Erro ao listar movimentações: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar movimentações")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": movements,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// GET /api/stock/expiring
func (h *StockHandler) GetExpiringLots(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	lots, err := h.stockService.GetExpiringLots(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar lotes com validade próxima: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar lotes com validade próxima")
		return
	}

	utils.SendSuccess(w, http.StatusOK, lots)
}

// GET /api/stock/low-stock
func (h *StockHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	products, err := h.stockService.GetLowStock(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar produtos com estoque baixo: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar produtos com estoque baixo")
		return
	}

	utils.SendSuccess(w, http.StatusOK, products)
}

// GET /api/stock/summary
func (h *StockHandler) GetStockSummary(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	summary, err := h.stockService.GetStockSummary(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar resumo de estoque: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar resumo de estoque")
		return
	}

	utils.SendSuccess(w, http.StatusOK, summary)
}