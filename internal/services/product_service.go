// ================================================================
// PACOTE SERVICES - PRODUCT SERVICE
// ================================================================
// ⚠️ CORRIGIDO: Create() não recebia nem setava associationID.
// Todas as funções agora recebem/filtram por associationID.
// ================================================================

package services

import (
	"errors"

	"cannacare-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

type CreateProductRequest struct {
	Name          string  `json:"name" validate:"required,min=3,max=200"`
	Description   string  `json:"description" validate:"omitempty"`
	UnitPrice     float64 `json:"unit_price" validate:"required,min=0"`
	MinStockAlert int     `json:"min_stock_alert" validate:"omitempty,min=0"`
}

type UpdateProductRequest struct {
	Name          string  `json:"name" validate:"omitempty,min=3,max=200"`
	Description   string  `json:"description" validate:"omitempty"`
	UnitPrice     float64 `json:"unit_price" validate:"omitempty,min=0"`
	MinStockAlert int     `json:"min_stock_alert" validate:"omitempty,min=0"`
	IsActive      *bool   `json:"is_active" validate:"omitempty"`
}

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

type ListProductRequest struct {
	Name     string `json:"name" query:"name"`
	IsActive *bool  `json:"is_active" query:"is_active"`
	Page     int    `json:"page" query:"page"`
	Limit    int    `json:"limit" query:"limit"`
}

// ================================================================
// FUNÇÃO CREATE()
// ================================================================
func (s *ProductService) Create(associationID uuid.UUID, req CreateProductRequest) (*ProductResponse, error) {
	// --- 1. Validar nome único DENTRO da associação ---
	var existingProduct models.Product
	err := s.db.Where("name = ? AND association_id = ?", req.Name, associationID).First(&existingProduct).Error
	if err == nil {
		return nil, errors.New("produto com este nome já existe")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	minStockAlert := req.MinStockAlert
	if minStockAlert == 0 {
		minStockAlert = 10
	}

	product := &models.Product{
		AssociationID: associationID, // ← ESSENCIAL!
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
func (s *ProductService) GetByID(associationID, id uuid.UUID) (*ProductResponse, error) {
	var product models.Product
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&product).Error; err != nil {
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
func (s *ProductService) List(associationID uuid.UUID, req ListProductRequest) ([]ProductResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	query := s.db.Model(&models.Product{}).Where("association_id = ?", associationID)

	if req.Name != "" {
		query = query.Where("name ILIKE ?", "%"+req.Name+"%")
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []models.Product
	if err := query.Offset(offset).Limit(req.Limit).Order("name ASC").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	var responses []ProductResponse
	for _, p := range products {
		responses = append(responses, *toProductResponse(&p))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATE()
// ================================================================
func (s *ProductService) Update(associationID, id uuid.UUID, req UpdateProductRequest) (*ProductResponse, error) {
	var product models.Product
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("produto não encontrado")
		}
		return nil, err
	}

	if req.Name != "" && req.Name != product.Name {
		var existingProduct models.Product
		err := s.db.Where("name = ? AND id != ? AND association_id = ?", req.Name, id, associationID).First(&existingProduct).Error
		if err == nil {
			return nil, errors.New("produto com este nome já existe")
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		product.Name = req.Name
	}

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
func (s *ProductService) Delete(associationID, id uuid.UUID) error {
	var count int64
	if err := s.db.Model(&models.PrescriptionItem{}).
		Where("product_id = ? AND association_id = ?", id, associationID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("produto está sendo usado em prescrições, não pode ser removido")
	}

	result := s.db.Where("association_id = ?", associationID).Delete(&models.Product{}, "id = ?", id)
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
func (s *ProductService) GetLowStock(associationID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_low_stock").Where("association_id = ?", associationID).Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ================================================================
// FUNÇÃO GETSTOCKSUMMARY()
// ================================================================
func (s *ProductService) GetStockSummary(associationID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_stock_summary").Where("association_id = ?", associationID).Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================
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