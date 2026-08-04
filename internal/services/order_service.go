// ================================================================
// PACOTE SERVICES - ORDER SERVICE
// ================================================================
// ⚠️ CORRIGIDO: Create() não recebia associationID. OrderItem e
// StockMovement criados dentro do fluxo também precisam do campo.
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

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

type CreateOrderRequest struct {
	PatientID      string                   `json:"patient_id" validate:"required"`
	PrescriptionID string                   `json:"prescription_id" validate:"required"`
	Items          []CreateOrderItemRequest `json:"items" validate:"required,min=1"`
	Notes          string                   `json:"notes" validate:"omitempty"`
}

type CreateOrderItemRequest struct {
	ProductLotID string  `json:"product_lot_id" validate:"required"`
	Quantity     int     `json:"quantity" validate:"required,min=1"`
	UnitPrice    float64 `json:"unit_price" validate:"required,min=0"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pendente separado dispensa correio entregue cancelado"`
	Notes  string `json:"notes" validate:"omitempty"`
}

type UpdateTrackingRequest struct {
	TrackingCode    string `json:"tracking_code" validate:"required"`
	ShippingCarrier string `json:"shipping_carrier" validate:"required"`
}

type OrderResponse struct {
	ID               string              `json:"id"`
	PatientID        string              `json:"patient_id"`
	PatientName      string              `json:"patient_name"`
	PrescriptionID   string              `json:"prescription_id"`
	Status           string              `json:"status"`
	TotalAmount      float64             `json:"total_amount"`
	Notes            string              `json:"notes,omitempty"`
	OrderDate        string              `json:"order_date"`
	StatusUpdatedAt  string              `json:"status_updated_at"`
	ShippingCarrier  string              `json:"shipping_carrier,omitempty"`
	TrackingCode     string              `json:"tracking_code,omitempty"`
	ShippingLabelURL string              `json:"shipping_label_url,omitempty"`
	ShippingCost     float64             `json:"shipping_cost,omitempty"`
	Items            []OrderItemResponse `json:"items"`
	CreatedAt        string              `json:"created_at"`
	UpdatedAt        string              `json:"updated_at"`
}

type OrderItemResponse struct {
	ID           string  `json:"id"`
	ProductLotID string  `json:"product_lot_id"`
	ProductName  string  `json:"product_name"`
	LotNumber    string  `json:"lot_number"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	TotalPrice   float64 `json:"total_price"`
}

type ListOrderRequest struct {
	PatientID      string `json:"patient_id" query:"patient_id"`
	PrescriptionID string `json:"prescription_id" query:"prescription_id"`
	Status         string `json:"status" query:"status"`
	Page           int    `json:"page" query:"page"`
	Limit          int    `json:"limit" query:"limit"`
}

type OrderValidationResult struct {
	IsValid bool   `json:"is_valid"`
	Message string `json:"message"`
	OrderID string `json:"order_id,omitempty"`
}

// ================================================================
// FUNÇÃO CREATE()
// ================================================================
func (s *OrderService) Create(associationID uuid.UUID, req CreateOrderRequest, userID uuid.UUID) (*OrderResponse, error) {
	// --- 1. Validar paciente (DENTRO da associação) ---
	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		return nil, fmt.Errorf("ID do paciente inválido: %w", err)
	}

	var patient models.Patient
	if err := s.db.Where("id = ? AND status = ? AND association_id = ?", patientID, "aprovado", associationID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado ou não aprovado")
		}
		return nil, err
	}

	// --- 2. Validar prescrição (DENTRO da associação) ---
	prescriptionID, err := uuid.Parse(req.PrescriptionID)
	if err != nil {
		return nil, fmt.Errorf("ID da prescrição inválido: %w", err)
	}

	var prescription models.Prescription
	if err := s.db.Where("id = ? AND is_active = ? AND status != ? AND association_id = ?", prescriptionID, true, "vencida", associationID).First(&prescription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("prescrição não encontrada, inativa ou vencida")
		}
		return nil, err
	}

	if prescription.PatientID != patientID {
		return nil, errors.New("prescrição não pertence a este paciente")
	}

	// --- 3. Validar estoque e criar itens (lotes DENTRO da associação) ---
	var orderItems []models.OrderItem
	var totalAmount float64

	for _, itemReq := range req.Items {
		lotID, err := uuid.Parse(itemReq.ProductLotID)
		if err != nil {
			return nil, fmt.Errorf("ID do lote inválido: %w", err)
		}

		var lot models.ProductLot
		if err := s.db.Preload("Product").Where("id = ? AND association_id = ?", lotID, associationID).First(&lot).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("lote não encontrado: %s", itemReq.ProductLotID)
			}
			return nil, err
		}

		if lot.CurrentQuantity < itemReq.Quantity {
			return nil, fmt.Errorf("estoque insuficiente para o lote %s. Disponível: %d, Solicitado: %d",
				lot.LotNumber, lot.CurrentQuantity, itemReq.Quantity)
		}

		if lot.ExpirationDate.Before(time.Now()) {
			return nil, fmt.Errorf("lote %s está vencido", lot.LotNumber)
		}

		var prescriptionItem models.PrescriptionItem
		if err := s.db.Where("prescription_id = ? AND product_id = ? AND association_id = ?", prescriptionID, lot.ProductID, associationID).First(&prescriptionItem).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				productName := ""
				if lot.Product != nil {
					productName = lot.Product.Name
				}
				return nil, fmt.Errorf("produto %s não está na prescrição", productName)
			}
			return nil, err
		}

		itemTotal := float64(itemReq.Quantity) * itemReq.UnitPrice
		orderItems = append(orderItems, models.OrderItem{
			AssociationID: associationID, // ← ESSENCIAL!
			ProductLotID:  lotID,
			Quantity:      itemReq.Quantity,
			UnitPrice:     itemReq.UnitPrice,
		})
		totalAmount += itemTotal
	}

	// --- 4. Criar o pedido (SEM os itens ainda) ---
	order := &models.Order{
		AssociationID:   associationID, // ← ESSENCIAL!
		PatientID:       patientID,
		PrescriptionID:  prescriptionID,
		Status:          "pendente",
		TotalAmount:     totalAmount,
		Notes:           req.Notes,
		OrderDate:       time.Now(),
		StatusUpdatedAt: time.Now(),
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}

	// --- 4b. Criar os itens SEPARADAMENTE, omitindo TotalPrice ---
	// ⚠️ total_price é uma coluna GERADA PELO POSTGRES (quantity * unit_price).
	// Se criarmos os itens junto com o pedido (via order.Items = orderItems),
	// o GORM ignora a permissão "só leitura" do campo e tenta inserir um
	// valor nela, o que o Postgres rejeita. Por isso criamos item por item,
	// com Omit("TotalPrice") explícito — assim o banco calcula sozinho.
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := s.db.Omit("TotalPrice").Create(&orderItems[i]).Error; err != nil {
			return nil, err
		}
	}
	order.Items = orderItems

	// --- 5. Dar baixa no estoque (atualizar quantidades) ---
	for _, item := range order.Items {
		var lot models.ProductLot
		if err := s.db.Where("id = ? AND association_id = ?", item.ProductLotID, associationID).First(&lot).Error; err != nil {
			return nil, err
		}

		previousQuantity := lot.CurrentQuantity
		lot.CurrentQuantity -= item.Quantity

		if err := s.db.Save(&lot).Error; err != nil {
			return nil, err
		}

		movement := &models.StockMovement{
			AssociationID:    associationID, // ← ESSENCIAL!
			ProductLotID:     lot.ID,
			OrderID:          &order.ID,
			UserID:           userID,
			Type:             "baixa_pedido",
			Quantity:         -item.Quantity,
			PreviousQuantity: previousQuantity,
			NewQuantity:      lot.CurrentQuantity,
			Notes:            fmt.Sprintf("Baixa por pedido %s", order.ID.String()),
		}

		if err := s.db.Create(movement).Error; err != nil {
			return nil, err
		}
	}

	// --- 6. Carregar relacionamentos para resposta ---
	if err := s.db.Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription").Where("id = ?", order.ID).First(order).Error; err != nil {
		return nil, err
	}

	return s.toOrderResponse(order), nil
}

// ================================================================
// FUNÇÃO GETBYID()
// ================================================================
func (s *OrderService) GetByID(associationID, id uuid.UUID) (*OrderResponse, error) {
	var order models.Order
	if err := s.db.Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription").Where("id = ? AND association_id = ?", id, associationID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("pedido não encontrado")
		}
		return nil, err
	}

	return s.toOrderResponse(&order), nil
}

// ================================================================
// FUNÇÃO LIST()
// ================================================================
func (s *OrderService) List(associationID uuid.UUID, req ListOrderRequest) ([]OrderResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	query := s.db.Model(&models.Order{}).Where("association_id = ?", associationID).
		Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription")

	if req.PatientID != "" {
		patientID, err := uuid.Parse(req.PatientID)
		if err == nil {
			query = query.Where("patient_id = ?", patientID)
		}
	}
	if req.PrescriptionID != "" {
		prescriptionID, err := uuid.Parse(req.PrescriptionID)
		if err == nil {
			query = query.Where("prescription_id = ?", prescriptionID)
		}
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []models.Order
	if err := query.Offset(offset).Limit(req.Limit).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	var responses []OrderResponse
	for _, order := range orders {
		responses = append(responses, *s.toOrderResponse(&order))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATESTATUS()
// ================================================================
func (s *OrderService) UpdateStatus(associationID, id uuid.UUID, req UpdateOrderStatusRequest, userID uuid.UUID) (*OrderResponse, error) {
	var order models.Order
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("pedido não encontrado")
		}
		return nil, err
	}

	if !s.isValidStatusTransition(order.Status, req.Status) {
		return nil, fmt.Errorf("transição de status inválida: %s → %s", order.Status, req.Status)
	}

	order.Status = req.Status
	order.StatusUpdatedAt = time.Now()
	if req.Notes != "" {
		order.Notes = req.Notes
	}

	if err := s.db.Save(&order).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription").Where("id = ?", order.ID).First(&order).Error; err != nil {
		return nil, err
	}

	return s.toOrderResponse(&order), nil
}

// ================================================================
// FUNÇÃO UPDATETRACKING()
// ================================================================
func (s *OrderService) UpdateTracking(associationID, id uuid.UUID, req UpdateTrackingRequest) (*OrderResponse, error) {
	var order models.Order
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("pedido não encontrado")
		}
		return nil, err
	}

	if order.Status != "correio" && order.Status != "entregue" {
		return nil, errors.New("pedido deve estar com status 'correio' ou 'entregue' para adicionar rastreio")
	}

	order.TrackingCode = req.TrackingCode
	order.ShippingCarrier = req.ShippingCarrier

	if err := s.db.Save(&order).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription").Where("id = ?", order.ID).First(&order).Error; err != nil {
		return nil, err
	}

	return s.toOrderResponse(&order), nil
}

// ================================================================
// FUNÇÃO GETBYPATIENT()
// ================================================================
func (s *OrderService) GetByPatient(associationID, patientID uuid.UUID) ([]OrderResponse, error) {
	var orders []models.Order
	if err := s.db.Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription").
		Where("patient_id = ? AND association_id = ?", patientID, associationID).Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, err
	}

	var responses []OrderResponse
	for _, order := range orders {
		responses = append(responses, *s.toOrderResponse(&order))
	}

	return responses, nil
}

// ================================================================
// FUNÇÃO GENERATELABEL()
// ================================================================
func (s *OrderService) GenerateLabel(associationID, id uuid.UUID) (string, error) {
	var order models.Order
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("pedido não encontrado")
		}
		return "", err
	}

	if order.Status != "dispensa" && order.Status != "correio" {
		return "", errors.New("pedido deve estar com status 'dispensa' ou 'correio' para gerar etiqueta")
	}

	labelURL := fmt.Sprintf("/api/orders/%s/label.pdf", order.ID.String())

	order.ShippingLabelURL = labelURL
	if err := s.db.Save(&order).Error; err != nil {
		return "", err
	}

	return labelURL, nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================
func (s *OrderService) isValidStatusTransition(currentStatus, newStatus string) bool {
	validTransitions := map[string][]string{
		"pendente":  {"separado", "cancelado"},
		"separado":  {"dispensa", "cancelado"},
		"dispensa":  {"correio", "cancelado"},
		"correio":   {"entregue", "cancelado"},
		"entregue":  {},
		"cancelado": {},
	}

	allowed, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}

func (s *OrderService) toOrderResponse(order *models.Order) *OrderResponse {
	items := []OrderItemResponse{}
	for _, item := range order.Items {
		productName := ""
		lotNumber := ""
		if item.ProductLot != nil && item.ProductLot.ID != uuid.Nil {
			lotNumber = item.ProductLot.LotNumber
			if item.ProductLot.Product != nil && item.ProductLot.Product.ID != uuid.Nil {
				productName = item.ProductLot.Product.Name
			}
		}
		items = append(items, OrderItemResponse{
			ID:           item.ID.String(),
			ProductLotID: item.ProductLotID.String(),
			ProductName:  productName,
			LotNumber:    lotNumber,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			TotalPrice:   item.TotalPrice,
		})
	}

	patientName := ""
	if order.Patient != nil && order.Patient.ID != uuid.Nil {
		patientName = order.Patient.FullName
	}

	return &OrderResponse{
		ID:               order.ID.String(),
		PatientID:        order.PatientID.String(),
		PatientName:      patientName,
		PrescriptionID:   order.PrescriptionID.String(),
		Status:           order.Status,
		TotalAmount:      order.TotalAmount,
		Notes:            order.Notes,
		OrderDate:        order.OrderDate.Format("2006-01-02 15:04:05"),
		StatusUpdatedAt:  order.StatusUpdatedAt.Format("2006-01-02 15:04:05"),
		ShippingCarrier:  order.ShippingCarrier,
		TrackingCode:     order.TrackingCode,
		ShippingLabelURL: order.ShippingLabelURL,
		ShippingCost:     order.ShippingCost,
		Items:            items,
		CreatedAt:        order.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        order.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}