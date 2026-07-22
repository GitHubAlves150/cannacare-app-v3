// ================================================================
// PACOTE SERVICES - PRODUCT SERVICE
// ================================================================
// Camada de serviço responsável pela gestão de produtos.
//
// RESPONSABILIDADES:
// 1. CRUD completo de produtos
// 2. Validação de dados (nome, preço, estoque mínimo)
// 3. Listagem com filtros e paginação
// 4. Controle de ativação/desativação
// 5. Busca por nome ou categoria
// ================================================================

package services

import (
	"errors"

	"cannacare-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT PRODUCTSERVICE
// ================================================================
type ProductService struct {
	db *gorm.DB
}

// ================================================================
// FUNÇÃO NEWPRODUCTSERVICE()
// ================================================================
func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

// CreateProductRequest - Dados para criar um novo produto
type CreateProductRequest struct {
	Name          string  `json:"name" validate:"required,min=3,max=200"`
	Description   string  `json:"description" validate:"omitempty"`
	UnitPrice     float64 `json:"unit_price" validate:"required,min=0"`
	MinStockAlert int     `json:"min_stock_alert" validate:"omitempty,min=0"`
}

// UpdateProductRequest - Dados para atualizar um produto
type UpdateProductRequest struct {
	Name          string  `json:"name" validate:"omitempty,min=3,max=200"`
	Description   string  `json:"description" validate:"omitempty"`
	UnitPrice     float64 `json:"unit_price" validate:"omitempty,min=0"`
	MinStockAlert int     `json:"min_stock_alert" validate:"omitempty,min=0"`
	IsActive      *bool   `json:"is_active" validate:"omitempty"`
}

// ProductResponse - Resposta com dados do produto
type ProductResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description,omitempty"`
	UnitPrice     float64 `json:"unit_price"`
	MinStockAlert int     `json:"min_stock_alert"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// ListProductRequest - Filtros para listagem
type ListProductRequest struct {
	Name     string `json:"name" query:"name"`
	IsActive *bool  `json:"is_active" query:"is_active"`
	Page     int    `json:"page" query:"page"`
	Limit    int    `json:"limit" query:"limit"`
}

// ================================================================
// FUNÇÃO CREATE()
// ================================================================
// Cria um novo produto no sistema
func (s *ProductService) Create(req CreateProductRequest) (*ProductResponse, error) {
	// --- 1. Validar nome único ---
	var existingProduct models.Product
	err := s.db.Where("name = ?", req.Name).First(&existingProduct).Error
	if err == nil {
		return nil, errors.New("produto com este nome já existe")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- 2. Definir estoque mínimo padrão ---
	minStockAlert := req.MinStockAlert
	if minStockAlert == 0 {
		minStockAlert = 10
	}

	// --- 3. Criar produto ---
	product := &models.Product{
		Name:          req.Name,
		Description:   req.Description,
		UnitPrice:     req.UnitPrice,
		MinStockAlert: minStockAlert,
		IsActive:      true,
	}

	if err := s.db.Create(product).Error; err != nil {
		return nil, err
	}

	return toProductResponse(product), nil
}

// ================================================================
// FUNÇÃO GETBYID()
// ================================================================
// Busca um produto pelo ID
func (s *ProductService) GetByID(id uuid.UUID) (*ProductResponse, error) {
	var product models.Product
	if err := s.db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("produto não encontrado")
		}
		return nil, err
	}
	return toProductResponse(&product), nil
}

// ================================================================
// FUNÇÃO LIST()
// ================================================================
// Lista produtos com filtros e paginação
func (s *ProductService) List(req ListProductRequest) ([]ProductResponse, int64, error) {
	// --- 1. Definir paginação ---
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	// --- 2. Construir query ---
	query := s.db.Model(&models.Product{})

	if req.Name != "" {
		query = query.Where("name ILIKE ?", "%"+req.Name+"%")
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// --- 3. Contar total ---
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// --- 4. Buscar com paginação ---
	var products []models.Product
	if err := query.Offset(offset).Limit(req.Limit).Order("name ASC").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	// --- 5. Converter para resposta ---
	var responses []ProductResponse
	for _, p := range products {
		responses = append(responses, *toProductResponse(&p))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATE()
// ================================================================
// Atualiza os dados de um produto
func (s *ProductService) Update(id uuid.UUID, req UpdateProductRequest) (*ProductResponse, error) {
	// --- 1. Buscar produto ---
	var product models.Product
	if err := s.db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("produto não encontrado")
		}
		return nil, err
	}

	// --- 2. Verificar nome duplicado (se for alterado) ---
	if req.Name != "" && req.Name != product.Name {
		var existingProduct models.Product
		err := s.db.Where("name = ? AND id != ?", req.Name, id).First(&existingProduct).Error
		if err == nil {
			return nil, errors.New("produto com este nome já existe")
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		product.Name = req.Name
	}

	// --- 3. Atualizar campos ---
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.UnitPrice > 0 {
		product.UnitPrice = req.UnitPrice
	}
	if req.MinStockAlert > 0 {
		product.MinStockAlert = req.MinStockAlert
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.db.Save(&product).Error; err != nil {
		return nil, err
	}

	return toProductResponse(&product), nil
}

// ================================================================
// FUNÇÃO DELETE()
// ================================================================
// Remove um produto (soft delete)
func (s *ProductService) Delete(id uuid.UUID) error {
	// Verificar se o produto não está sendo usado em prescrições
	var count int64
	if err := s.db.Model(&models.PrescriptionItem{}).Where("product_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("produto está sendo usado em prescrições, não pode ser removido")
	}

	result := s.db.Delete(&models.Product{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("produto não encontrado")
	}
	return nil
}

// ================================================================
// FUNÇÃO GETLOWSTOCK()
// ================================================================
// Retorna produtos com estoque baixo (usando view)
func (s *ProductService) GetLowStock() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_low_stock").Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ================================================================
// FUNÇÃO GETSTOCKSUMMARY()
// ================================================================
// Retorna resumo de estoque (usando view)
func (s *ProductService) GetStockSummary() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_stock_summary").Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// toProductResponse - Converte models.Product para ProductResponse
func toProductResponse(product *models.Product) *ProductResponse {
	return &ProductResponse{
		ID:            product.ID.String(),
		Name:          product.Name,
		Description:   product.Description,
		UnitPrice:     product.UnitPrice,
		MinStockAlert: product.MinStockAlert,
		IsActive:      product.IsActive,
		CreatedAt:     product.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     product.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
