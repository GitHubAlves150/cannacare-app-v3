// ================================================================
// PACOTE SERVICES - STOCK SERVICE
// ================================================================
// Camada de serviço responsável pelo controle de estoque.
//
// RESPONSABILIDADES:
// 1. CRUD de lotes de produtos
// 2. Registrar entrada de produtos
// 3. Registrar saída de produtos (baixa por pedido)
// 4. Ajustes manuais de estoque
// 5. Alertas de validade próxima
// 6. Histórico de movimentações
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

// ================================================================
// STRUCT STOCKSERVICE
// ================================================================
type StockService struct {
	db *gorm.DB
}

// ================================================================
// FUNÇÃO NEWSTOCKSERVICE()
// ================================================================
func NewStockService(db *gorm.DB) *StockService {
	return &StockService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

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
// Cria um novo lote para um produto
func (s *StockService) CreateLot(req CreateLotRequest, userID uuid.UUID) (*LotResponse, error) {
	// --- 1. Validar produto ---
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("ID do produto inválido: %w", err)
	}

	var product models.Product
	if err := s.db.Where("id = ?", productID).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("produto não encontrado")
		}
		return nil, err
	}

	// --- 2. Validar data de validade ---
	if req.ExpirationDate.Before(time.Now()) {
		return nil, errors.New("data de validade não pode ser anterior à data atual")
	}

	// --- 3. Validar lote único por produto ---
	var existingLot models.ProductLot
	err = s.db.Where("product_id = ? AND lot_number = ?", productID, req.LotNumber).First(&existingLot).Error
	if err == nil {
		return nil, fmt.Errorf("lote %s já existe para este produto", req.LotNumber)
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- 4. Criar lote ---
	now := time.Now()
	lot := &models.ProductLot{
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
// Ajusta manualmente o estoque de um lote
func (s *StockService) AdjustStock(req StockAdjustRequest, userID uuid.UUID) (*StockMovementResponse, error) {
	// --- 1. Validar lote ---
	lotID, err := uuid.Parse(req.ProductLotID)
	if err != nil {
		return nil, fmt.Errorf("ID do lote inválido: %w", err)
	}

	var lot models.ProductLot
	if err := s.db.Where("id = ?", lotID).First(&lot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("lote não encontrado")
		}
		return nil, err
	}

	// --- 2. Validar quantidade ---
	if req.Quantity < 0 && lot.CurrentQuantity+req.Quantity < 0 {
		return nil, fmt.Errorf("estoque insuficiente. Disponível: %d", lot.CurrentQuantity)
	}

	// --- 3. Atualizar quantidade ---
	previousQuantity := lot.CurrentQuantity
	lot.CurrentQuantity += req.Quantity

	if err := s.db.Save(&lot).Error; err != nil {
		return nil, err
	}

	// --- 4. Registrar movimentação ---
	movementType := "ajuste_manual"
	if req.Quantity < 0 {
		movementType = "perda"
	}

	movement := &models.StockMovement{
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
// Busca um lote pelo ID
func (s *StockService) GetLotByID(id uuid.UUID) (*LotResponse, error) {
	var lot models.ProductLot
	if err := s.db.Preload("Product").Where("id = ?", id).First(&lot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("lote não encontrado")
		}
		return nil, err
	}

	// ✅ CORRETO: passar o produto como valor (não ponteiro)
	return s.toLotResponse(&lot, lot.Product), nil
}

// ================================================================
// FUNÇÃO LISTLOTS()
// ================================================================
// Lista lotes com filtros e paginação
func (s *StockService) ListLots(req ListLotRequest) ([]LotResponse, int64, error) {
	// --- 1. Definir paginação ---
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	// --- 2. Construir query ---
	query := s.db.Model(&models.ProductLot{}).Preload("Product")

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

	// --- 3. Contar total ---
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// --- 4. Buscar com paginação ---
	var lots []models.ProductLot
	if err := query.Offset(offset).Limit(req.Limit).Order("created_at DESC").Find(&lots).Error; err != nil {
		return nil, 0, err
	}

	// --- 5. Converter para resposta ---
	var responses []LotResponse
	for i := range lots {
		responses = append(responses, *s.toLotResponse(&lots[i], lots[i].Product))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO GETMOVEMENTS()
// ================================================================
// Lista movimentações de estoque com filtros
func (s *StockService) GetMovements(req ListMovementRequest) ([]StockMovementResponse, int64, error) {
	// --- 1. Definir paginação ---
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	// --- 2. Construir query ---
	query := s.db.Model(&models.StockMovement{}).
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

	// --- 3. Contar total ---
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// --- 4. Buscar com paginação ---
	var movements []models.StockMovement
	if err := query.Offset(offset).Limit(req.Limit).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, 0, err
	}

	// --- 5. Converter para resposta ---
	var responses []StockMovementResponse
	for i := range movements {
		responses = append(responses, *s.toMovementResponse(&movements[i]))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO GETEXPIRINGLOTS()
// ================================================================
// Retorna lotes com validade próxima (30 dias)
func (s *StockService) GetExpiringLots() ([]LotResponse, error) {
	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)

	var lots []models.ProductLot
	if err := s.db.Preload("Product").
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
// Retorna produtos com estoque baixo (usando view)
func (s *StockService) GetLowStock() ([]map[string]interface{}, error) {
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
func (s *StockService) GetStockSummary() ([]map[string]interface{}, error) {
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

// toLotResponse - Converte models.ProductLot para LotResponse
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

// toMovementResponse - Converte models.StockMovement para StockMovementResponse
func (s *StockService) toMovementResponse(movement *models.StockMovement) *StockMovementResponse {
	productName := ""
	lotNumber := ""
	if movement.ProductLot.ID != uuid.Nil {
		if movement.ProductLot.Product.ID != uuid.Nil {
			productName = movement.ProductLot.Product.Name
		}
		lotNumber = movement.ProductLot.LotNumber
	}

	userName := ""
	if movement.User.ID != uuid.Nil {
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
