// ================================================================
// PACOTE HANDLERS - PRODUCT HANDLER
// ================================================================
// Camada HTTP que lida com as requisições de produtos.
//
// ENDPOINTS:
//   POST   /api/products           - Criar produto
//   GET    /api/products           - Listar produtos (filtros)
//   GET    /api/products/{id}      - Buscar produto por ID
//   PUT    /api/products/{id}      - Atualizar produto
//   DELETE /api/products/{id}      - Remover produto (soft delete)
//   GET    /api/products/low-stock - Produtos com estoque baixo
//   GET    /api/products/stock-summary - Resumo de estoque
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

// ================================================================
// STRUCT PRODUCTHANDLER
// ================================================================
type ProductHandler struct {
	productService *services.ProductService
	validator      *validator.Validate
}

// ================================================================
// FUNÇÃO NEWPRODUCTHANDLER()
// ================================================================
func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		validator:      validator.New(),
	}
}

// ================================================================
// HANDLER: CREATE
// ================================================================
// Endpoint: POST /api/products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req services.CreateProductRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	product, err := h.productService.Create(req)
	if err != nil {
		log.Printf("❌ Erro ao criar produto: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, product)
}

// ================================================================
// HANDLER: GET BY ID
// ================================================================
// Endpoint: GET /api/products/{id}
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	product, err := h.productService.GetByID(id)
	if err != nil {
		if err.Error() == "produto não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar produto: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar produto")
		return
	}

	utils.SendSuccess(w, http.StatusOK, product)
}

// ================================================================
// HANDLER: LIST
// ================================================================
// Endpoint: GET /api/products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	req := services.ListProductRequest{
		Name: r.URL.Query().Get("name"),
	}

	if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			req.IsActive = &isActive
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

	products, total, err := h.productService.List(req)
	if err != nil {
		log.Printf("❌ Erro ao listar produtos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar produtos")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": products,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// ================================================================
// HANDLER: UPDATE
// ================================================================
// Endpoint: PUT /api/products/{id}
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req services.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	product, err := h.productService.Update(id, req)
	if err != nil {
		if err.Error() == "produto não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar produto: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, product)
}

// ================================================================
// HANDLER: DELETE
// ================================================================
// Endpoint: DELETE /api/products/{id}
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.productService.Delete(id); err != nil {
		if err.Error() == "produto não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao deletar produto: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Produto removido com sucesso",
	})
}

// ================================================================
// HANDLER: GET LOW STOCK
// ================================================================
// Endpoint: GET /api/products/low-stock
func (h *ProductHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	products, err := h.productService.GetLowStock()
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
// Endpoint: GET /api/products/stock-summary
func (h *ProductHandler) GetStockSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.productService.GetStockSummary()
	if err != nil {
		log.Printf("❌ Erro ao buscar resumo de estoque: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar resumo de estoque")
		return
	}

	utils.SendSuccess(w, http.StatusOK, summary)
}
