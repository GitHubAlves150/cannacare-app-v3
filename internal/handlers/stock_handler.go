// ================================================================
// PACOTE HANDLERS - STOCK HANDLER
// ================================================================
// Camada HTTP que lida com as requisições de estoque.
//
// ENDPOINTS:
//   POST   /api/stock/lots           - Criar lote
//   GET    /api/stock/lots           - Listar lotes (filtros)
//   GET    /api/stock/lots/{id}      - Buscar lote por ID
//   POST   /api/stock/adjust         - Ajustar estoque manualmente
//   GET    /api/stock/movements      - Listar movimentações
//   GET    /api/stock/expiring       - Lotes com validade próxima
//   GET    /api/stock/low-stock      - Produtos com estoque baixo
//   GET    /api/stock/summary        - Resumo de estoque
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

// ================================================================
// STRUCT STOCKHANDLER
// ================================================================
type StockHandler struct {
	stockService *services.StockService
	validator    *validator.Validate
}

// ================================================================
// FUNÇÃO NEWSTOCKHANDLER()
// ================================================================
func NewStockHandler(stockService *services.StockService) *StockHandler {
	return &StockHandler{
		stockService: stockService,
		validator:    validator.New(),
	}
}

// ================================================================
// HANDLER: CREATE LOT
// ================================================================
// Endpoint: POST /api/stock/lots
func (h *StockHandler) CreateLot(w http.ResponseWriter, r *http.Request) {
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

	lot, err := h.stockService.CreateLot(req, userID)
	if err != nil {
		log.Printf("❌ Erro ao criar lote: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, lot)
}

// ================================================================
// HANDLER: GET LOT BY ID
// ================================================================
// Endpoint: GET /api/stock/lots/{id}
func (h *StockHandler) GetLotByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	lot, err := h.stockService.GetLotByID(id)
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

// ================================================================
// HANDLER: LIST LOTS
// ================================================================
// Endpoint: GET /api/stock/lots
func (h *StockHandler) ListLots(w http.ResponseWriter, r *http.Request) {
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

	lots, total, err := h.stockService.ListLots(req)
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

// ================================================================
// HANDLER: ADJUST STOCK
// ================================================================
// Endpoint: POST /api/stock/adjust
func (h *StockHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
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

	movement, err := h.stockService.AdjustStock(req, userID)
	if err != nil {
		log.Printf("❌ Erro ao ajustar estoque: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, movement)
}

// ================================================================
// HANDLER: GET MOVEMENTS
// ================================================================
// Endpoint: GET /api/stock/movements
func (h *StockHandler) GetMovements(w http.ResponseWriter, r *http.Request) {
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

	movements, total, err := h.stockService.GetMovements(req)
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

// ================================================================
// HANDLER: GET EXPIRING LOTS
// ================================================================
// Endpoint: GET /api/stock/expiring
func (h *StockHandler) GetExpiringLots(w http.ResponseWriter, r *http.Request) {
	lots, err := h.stockService.GetExpiringLots()
	if err != nil {
		log.Printf("❌ Erro ao buscar lotes com validade próxima: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar lotes com validade próxima")
		return
	}

	utils.SendSuccess(w, http.StatusOK, lots)
}

// ================================================================
// HANDLER: GET LOW STOCK
// ================================================================
// Endpoint: GET /api/stock/low-stock
func (h *StockHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	products, err := h.stockService.GetLowStock()
	if err != nil {
		log.Printf("❌ Erro ao buscar produtos com estoque baixo: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar produtos com estoque baixo")
		return
	}

	utils.SendSuccess(w, http.StatusOK, products)
}

// ================================================================
// HANDLER: GET STOCK SUMMARY
// ================================================================
// Endpoint: GET /api/stock/summary
func (h *StockHandler) GetStockSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.stockService.GetStockSummary()
	if err != nil {
		log.Printf("❌ Erro ao buscar resumo de estoque: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar resumo de estoque")
		return
	}

	utils.SendSuccess(w, http.StatusOK, summary)
}
