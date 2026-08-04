// ================================================================
// PACOTE SERVICES - STOCK SERVICE
// ================================================================
// ⚠️ CORRIGIDO: CreateLot() e AdjustStock() não setavam associationID
// no ProductLot/StockMovement criados. Todas as buscas agora também
// filtram por associação.
// ================================================================

package services

import (
	"errors"
	"fmt"
	"time"

	"cannacare-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StockService struct {
	db *gorm.DB
}

func NewStockService(db *gorm.DB) *StockService {
	return &StockService{db: db}
}

type CreateLotRequest struct {
	ProductID      string    `json:"product_id" validate:"required"`
	LotNumber      string    `json:"lot_number" validate:"required"`
	ExpirationDate time.Time `json:"expiration_date" validate:"required"`
	Quantity       int       `json:"quantity" validate:"required,min=1"`
	Supplier       string    `json:"supplier" validate:"omitempty"`
	PurchaseDate   time.Time `json:"purchase_date" validate:"omitempty"`
	PurchasePrice  float64   `json:"purchase_price" validate:"omitempty,min=0"`
}

type UpdateLotRequest struct {
	LotNumber      string    `json:"lot_number" validate:"omitempty"`
	ExpirationDate time.Time `json:"expiration_date" validate:"omitempty"`
	Supplier       string    `json:"supplier" validate:"omitempty"`
	PurchaseDate   time.Time `json:"purchase_date" validate:"omitempty"`
	PurchasePrice  float64   `json:"purchase_price" validate:"omitempty,min=0"`
}

type StockMovementRequest struct {
	ProductLotID string `json:"product_lot_id" validate:"required"`
	Quantity     int    `json:"quantity" validate:"required,min=1"`
	Notes        string `json:"notes" validate:"omitempty"`
}

type StockAdjustRequest struct {
	ProductLotID string `json:"product_lot_id" validate:"required"`
	Quantity     int    `json:"quantity" validate:"required"`
	Reason       string `json:"reason" validate:"required"`
}

type LotResponse struct {
	ID              string  `json:"id"`
	ProductID       string  `json:"product_id"`
	ProductName     string  `json:"product_name"`
	LotNumber       string  `json:"lot_number"`
	ExpirationDate  string  `json:"expiration_date"`
	CurrentQuantity int     `json:"current_quantity"`
	InitialQuantity int     `json:"initial_quantity"`
	Supplier        string  `json:"supplier,omitempty"`
	PurchaseDate    string  `json:"purchase_date,omitempty"`
	PurchasePrice   float64 `json:"purchase_price,omitempty"`
	ReceivedAt      string  `json:"received_at,omitempty"`
	IsExpired       bool    `json:"is_expired"`
	DaysUntilExpire int     `json:"days_until_expire"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type StockMovementResponse struct {
	ID               string `json:"id"`
	ProductLotID     string `json:"product_lot_id"`
	ProductName      string `json:"product_name"`
	LotNumber        string `json:"lot_number"`
	OrderID          string `json:"order_id,omitempty"`
	UserID           string `json:"user_id"`
	UserName         string `json:"user_name"`
	Type             string `json:"type"`
	Quantity         int    `json:"quantity"`
	PreviousQuantity int    `json:"previous_quantity"`
	NewQuantity      int    `json:"new_quantity"`
	Notes            string `json:"notes,omitempty"`
	CreatedAt        string `json:"created_at"`
}

type ListLotRequest struct {
	ProductID string `json:"product_id" query:"product_id"`
	IsExpired *bool  `json:"is_expired" query:"is_expired"`
	Page      int    `json:"page" query:"page"`
	Limit     int    `json:"limit" query:"limit"`
}

type ListMovementRequest struct {
	ProductLotID string `json:"product_lot_id" query:"product_lot_id"`
	Type         string `json:"type" query:"type"`
	Page         int    `json:"page" query:"page"`
	Limit        int    `json:"limit" query:"limit"`
}

// ================================================================
// FUNÇÃO CREATELOT()
// ================================================================
func (s *StockService) CreateLot(associationID uuid.UUID, req CreateLotRequest, userID uuid.UUID) (*LotResponse, error) {
	// --- 1. Validar produto (DENTRO da associação) ---
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("ID do produto inválido: %w", err)
	}

	var product models.Product
	if err := s.db.Where("id = ? AND association_id = ?", productID, associationID).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("produto não encontrado")
		}
		return nil, err
	}

	// --- 2. Validar data de validade ---
	if req.ExpirationDate.Before(time.Now()) {
		return nil, errors.New("data de validade não pode ser anterior à data atual")
	}

	// --- 3. Validar lote único por produto (DENTRO da associação) ---
	var existingLot models.ProductLot
	err = s.db.Where("product_id = ? AND lot_number = ? AND association_id = ?", productID, req.LotNumber, associationID).First(&existingLot).Error
	if err == nil {
		return nil, fmt.Errorf("lote %s já existe para este produto", req.LotNumber)
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- 4. Criar lote ---
	now := time.Now()
	lot := &models.ProductLot{
		AssociationID:   associationID, // ← ESSENCIAL!
		ProductID:       productID,
		LotNumber:       req.LotNumber,
		ExpirationDate:  req.ExpirationDate,
		CurrentQuantity: req.Quantity,
		InitialQuantity: req.Quantity,
		Supplier:        req.Supplier,
		PurchaseDate:    &req.PurchaseDate,
		PurchasePrice:   req.PurchasePrice,
		ReceivedBy:      &userID,
		ReceivedAt:      &now,
	}

	if err := s.db.Create(lot).Error; err != nil {
		return nil, err
	}

	// --- 5. Registrar movimentação de entrada ---
	movement := &models.StockMovement{
		AssociationID:    associationID, // ← ESSENCIAL!
		ProductLotID:     lot.ID,
		UserID:           userID,
		Type:             "entrada",
		Quantity:         req.Quantity,
		PreviousQuantity: 0,
		NewQuantity:      req.Quantity,
		Notes:            "Entrada de produto - Lote " + req.LotNumber,
	}

	if err := s.db.Create(movement).Error; err != nil {
		return nil, err
	}

	return s.toLotResponse(lot, &product), nil
}

// ================================================================
// FUNÇÃO ADJUSTSTOCK()
// ================================================================
func (s *StockService) AdjustStock(associationID uuid.UUID, req StockAdjustRequest, userID uuid.UUID) (*StockMovementResponse, error) {
	lotID, err := uuid.Parse(req.ProductLotID)
	if err != nil {
		return nil, fmt.Errorf("ID do lote inválido: %w", err)
	}

	var lot models.ProductLot
	if err := s.db.Where("id = ? AND association_id = ?", lotID, associationID).First(&lot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("lote não encontrado")
		}
		return nil, err
	}

	if req.Quantity < 0 && lot.CurrentQuantity+req.Quantity < 0 {
		return nil, fmt.Errorf("estoque insuficiente. Disponível: %d", lot.CurrentQuantity)
	}

	previousQuantity := lot.CurrentQuantity
	lot.CurrentQuantity += req.Quantity

	if err := s.db.Save(&lot).Error; err != nil {
		return nil, err
	}

	movementType := "ajuste_manual"
	if req.Quantity < 0 {
		movementType = "perda"
	}

	movement := &models.StockMovement{
		AssociationID:    associationID, // ← ESSENCIAL!
		ProductLotID:     lotID,
		UserID:           userID,
		Type:             movementType,
		Quantity:         req.Quantity,
		PreviousQuantity: previousQuantity,
		NewQuantity:      lot.CurrentQuantity,
		Notes:            req.Reason,
	}

	if err := s.db.Create(movement).Error; err != nil {
		return nil, err
	}

	return s.toMovementResponse(movement), nil
}

// ================================================================
// FUNÇÃO GETLOTBYID()
// ================================================================
func (s *StockService) GetLotByID(associationID, id uuid.UUID) (*LotResponse, error) {
	var lot models.ProductLot
	if err := s.db.Preload("Product").Where("id = ? AND association_id = ?", id, associationID).First(&lot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("lote não encontrado")
		}
		return nil, err
	}

	return s.toLotResponse(&lot, lot.Product), nil
}

// ================================================================
// FUNÇÃO LISTLOTS()
// ================================================================
func (s *StockService) ListLots(associationID uuid.UUID, req ListLotRequest) ([]LotResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	query := s.db.Model(&models.ProductLot{}).Where("association_id = ?", associationID).Preload("Product")

	if req.ProductID != "" {
		productID, err := uuid.Parse(req.ProductID)
		if err == nil {
			query = query.Where("product_id = ?", productID)
		}
	}
	if req.IsExpired != nil {
		if *req.IsExpired {
			query = query.Where("expiration_date < ?", time.Now())
		} else {
			query = query.Where("expiration_date >= ?", time.Now())
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var lots []models.ProductLot
	if err := query.Offset(offset).Limit(req.Limit).Order("created_at DESC").Find(&lots).Error; err != nil {
		return nil, 0, err
	}

	var responses []LotResponse
	for i := range lots {
		responses = append(responses, *s.toLotResponse(&lots[i], lots[i].Product))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO GETMOVEMENTS()
// ================================================================
func (s *StockService) GetMovements(associationID uuid.UUID, req ListMovementRequest) ([]StockMovementResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	query := s.db.Model(&models.StockMovement{}).Where("association_id = ?", associationID).
		Preload("ProductLot").
		Preload("ProductLot.Product").
		Preload("User")

	if req.ProductLotID != "" {
		lotID, err := uuid.Parse(req.ProductLotID)
		if err == nil {
			query = query.Where("product_lot_id = ?", lotID)
		}
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var movements []models.StockMovement
	if err := query.Offset(offset).Limit(req.Limit).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, 0, err
	}

	var responses []StockMovementResponse
	for i := range movements {
		responses = append(responses, *s.toMovementResponse(&movements[i]))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO GETEXPIRINGLOTS()
// ================================================================
func (s *StockService) GetExpiringLots(associationID uuid.UUID) ([]LotResponse, error) {
	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)

	var lots []models.ProductLot
	if err := s.db.Preload("Product").
		Where("association_id = ?", associationID).
		Where("expiration_date BETWEEN ? AND ?", time.Now(), thirtyDaysFromNow).
		Where("current_quantity > 0").
		Order("expiration_date ASC").
		Find(&lots).Error; err != nil {
		return nil, err
	}

	var responses []LotResponse
	for i := range lots {
		responses = append(responses, *s.toLotResponse(&lots[i], lots[i].Product))
	}

	return responses, nil
}

// ================================================================
// FUNÇÃO GETLOWSTOCK()
// ================================================================
func (s *StockService) GetLowStock(associationID uuid.UUID) ([]map[string]interface{}, error) {
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
func (s *StockService) GetStockSummary(associationID uuid.UUID) ([]map[string]interface{}, error) {
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
func (s *StockService) toLotResponse(lot *models.ProductLot, product *models.Product) *LotResponse {
	daysUntilExpire := int(lot.ExpirationDate.Sub(time.Now()).Hours() / 24)
	isExpired := daysUntilExpire < 0

	purchaseDate := ""
	if lot.PurchaseDate != nil {
		purchaseDate = lot.PurchaseDate.Format("2006-01-02")
	}
	receivedAt := ""
	if lot.ReceivedAt != nil {
		receivedAt = lot.ReceivedAt.Format("2006-01-02 15:04:05")
	}

	productName := ""
	if product != nil && product.ID != uuid.Nil {
		productName = product.Name
	}

	return &LotResponse{
		ID:              lot.ID.String(),
		ProductID:       lot.ProductID.String(),
		ProductName:     productName,
		LotNumber:       lot.LotNumber,
		ExpirationDate:  lot.ExpirationDate.Format("2006-01-02"),
		CurrentQuantity: lot.CurrentQuantity,
		InitialQuantity: lot.InitialQuantity,
		Supplier:        lot.Supplier,
		PurchaseDate:    purchaseDate,
		PurchasePrice:   lot.PurchasePrice,
		ReceivedAt:      receivedAt,
		IsExpired:       isExpired,
		DaysUntilExpire: daysUntilExpire,
		CreatedAt:       lot.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       lot.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *StockService) toMovementResponse(movement *models.StockMovement) *StockMovementResponse {
	productName := ""
	lotNumber := ""
	if movement.ProductLot != nil && movement.ProductLot.ID != uuid.Nil {
		if movement.ProductLot.Product != nil && movement.ProductLot.Product.ID != uuid.Nil {
			productName = movement.ProductLot.Product.Name
		}
		lotNumber = movement.ProductLot.LotNumber
	}

	userName := ""
	if movement.User != nil && movement.User.ID != uuid.Nil {
		userName = movement.User.Name
	}

	orderID := ""
	if movement.OrderID != nil {
		orderID = movement.OrderID.String()
	}

	return &StockMovementResponse{
		ID:               movement.ID.String(),
		ProductLotID:     movement.ProductLotID.String(),
		ProductName:      productName,
		LotNumber:        lotNumber,
		OrderID:          orderID,
		UserID:           movement.UserID.String(),
		UserName:         userName,
		Type:             movement.Type,
		Quantity:         movement.Quantity,
		PreviousQuantity: movement.PreviousQuantity,
		NewQuantity:      movement.NewQuantity,
		Notes:            movement.Notes,
		CreatedAt:        movement.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}