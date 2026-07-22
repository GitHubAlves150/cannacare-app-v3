// ================================================================
// PACOTE SERVICES - ORDER SERVICE
// ================================================================
// Camada de serviço responsável pela gestão de pedidos.
//
// RESPONSABILIDADES:
// 1. CRUD completo de pedidos
// 2. Validar receita antes de criar pedido
// 3. Validar estoque disponível
// 4. Baixa automática no estoque
// 5. Controle de status do pedido
// 6. Geração de etiqueta de envio (PDF)
// 7. Histórico de pedidos por paciente
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
// STRUCT ORDERSERVICE
// ================================================================
type OrderService struct {
	db *gorm.DB
}

// ================================================================
// FUNÇÃO NEWORDERSERVICE()
// ================================================================
func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

// CreateOrderRequest - Dados para criar um novo pedido
type CreateOrderRequest struct {
	PatientID      string                        `json:"patient_id" validate:"required"`
	PrescriptionID string                        `json:"prescription_id" validate:"required"`
	Items          []CreateOrderItemRequest      `json:"items" validate:"required,min=1"`
	Notes          string                        `json:"notes" validate:"omitempty"`
}

// CreateOrderItemRequest - Item do pedido
type CreateOrderItemRequest struct {
	ProductLotID string  `json:"product_lot_id" validate:"required"`
	Quantity     int     `json:"quantity" validate:"required,min=1"`
	UnitPrice    float64 `json:"unit_price" validate:"required,min=0"`
}

// UpdateOrderStatusRequest - Para atualizar status do pedido
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pendente separado dispensa correio entregue cancelado"`
	Notes  string `json:"notes" validate:"omitempty"`
}

// UpdateTrackingRequest - Para adicionar código de rastreio
type UpdateTrackingRequest struct {
	TrackingCode   string `json:"tracking_code" validate:"required"`
	ShippingCarrier string `json:"shipping_carrier" validate:"required"`
}

// OrderResponse - Resposta com dados do pedido
type OrderResponse struct {
	ID               string                   `json:"id"`
	PatientID        string                   `json:"patient_id"`
	PatientName      string                   `json:"patient_name"`
	PrescriptionID   string                   `json:"prescription_id"`
	Status           string                   `json:"status"`
	TotalAmount      float64                  `json:"total_amount"`
	Notes            string                   `json:"notes,omitempty"`
	OrderDate        string                   `json:"order_date"`
	StatusUpdatedAt  string                   `json:"status_updated_at"`
	ShippingCarrier  string                   `json:"shipping_carrier,omitempty"`
	TrackingCode     string                   `json:"tracking_code,omitempty"`
	ShippingLabelURL string                   `json:"shipping_label_url,omitempty"`
	ShippingCost     float64                  `json:"shipping_cost,omitempty"`
	Items            []OrderItemResponse      `json:"items"`
	CreatedAt        string                   `json:"created_at"`
	UpdatedAt        string                   `json:"updated_at"`
}

// OrderItemResponse - Item do pedido
type OrderItemResponse struct {
	ID           string  `json:"id"`
	ProductLotID string  `json:"product_lot_id"`
	ProductName  string  `json:"product_name"`
	LotNumber    string  `json:"lot_number"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	TotalPrice   float64 `json:"total_price"`
}

// ListOrderRequest - Filtros para listagem
type ListOrderRequest struct {
	PatientID      string `json:"patient_id" query:"patient_id"`
	PrescriptionID string `json:"prescription_id" query:"prescription_id"`
	Status         string `json:"status" query:"status"`
	Page           int    `json:"page" query:"page"`
	Limit          int    `json:"limit" query:"limit"`
}

// OrderValidationResult - Resultado da validação do pedido
type OrderValidationResult struct {
	IsValid   bool   `json:"is_valid"`
	Message   string `json:"message"`
	OrderID   string `json:"order_id,omitempty"`
}

// ================================================================
// FUNÇÃO CREATE()
// ================================================================
// Cria um novo pedido
//
// VALIDAÇÕES:
// 1. Paciente existe e está ativo
// 2. Prescrição existe e é válida
// 3. Estoque disponível para cada item
// 4. Dar baixa no estoque
func (s *OrderService) Create(req CreateOrderRequest, userID uuid.UUID) (*OrderResponse, error) {
	// --- 1. Validar paciente ---
	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		return nil, fmt.Errorf("ID do paciente inválido: %w", err)
	}

	var patient models.Patient
	if err := s.db.Where("id = ? AND status = ?", patientID, "aprovado").First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado ou não aprovado")
		}
		return nil, err
	}

	// --- 2. Validar prescrição ---
	prescriptionID, err := uuid.Parse(req.PrescriptionID)
	if err != nil {
		return nil, fmt.Errorf("ID da prescrição inválido: %w", err)
	}

	var prescription models.Prescription
	if err := s.db.Where("id = ? AND is_active = ? AND status != ?", prescriptionID, true, "vencida").First(&prescription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("prescrição não encontrada, inativa ou vencida")
		}
		return nil, err
	}

	// Verificar se a prescrição pertence ao paciente
	if prescription.PatientID != patientID {
		return nil, errors.New("prescrição não pertence a este paciente")
	}

	// --- 3. Validar estoque e criar itens ---
	var orderItems []models.OrderItem
	var totalAmount float64

	for _, itemReq := range req.Items {
		lotID, err := uuid.Parse(itemReq.ProductLotID)
		if err != nil {
			return nil, fmt.Errorf("ID do lote inválido: %w", err)
		}

		// Buscar lote
		var lot models.ProductLot
		if err := s.db.Preload("Product").Where("id = ?", lotID).First(&lot).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("lote não encontrado: %s", itemReq.ProductLotID)
			}
			return nil, err
		}

		// Validar estoque
		if lot.CurrentQuantity < itemReq.Quantity {
			return nil, fmt.Errorf("estoque insuficiente para o lote %s. Disponível: %d, Solicitado: %d",
				lot.LotNumber, lot.CurrentQuantity, itemReq.Quantity)
		}

		// Validar validade
		if lot.ExpirationDate.Before(time.Now()) {
			return nil, fmt.Errorf("lote %s está vencido", lot.LotNumber)
		}

		// Validar se o lote pertence a um produto da prescrição
		// (verificar se o produto está na prescrição)
		var prescriptionItem models.PrescriptionItem
		if err := s.db.Where("prescription_id = ? AND product_id = ?", prescriptionID, lot.ProductID).First(&prescriptionItem).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("produto %s não está na prescrição", lot.Product.Name)
			}
			return nil, err
		}

		// Criar item do pedido
		itemTotal := float64(itemReq.Quantity) * itemReq.UnitPrice
		orderItems = append(orderItems, models.OrderItem{
			ProductLotID: lotID,
			Quantity:     itemReq.Quantity,
			UnitPrice:    itemReq.UnitPrice,
		})
		totalAmount += itemTotal
	}

	// --- 4. Criar o pedido ---
	order := &models.Order{
		PatientID:      patientID,
		PrescriptionID: prescriptionID,
		Status:         "pendente",
		TotalAmount:    totalAmount,
		Notes:          req.Notes,
		OrderDate:      time.Now(),
		StatusUpdatedAt: time.Now(),
		Items:          orderItems,
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}

	// --- 5. Dar baixa no estoque (atualizar quantidades) ---
	for _, item := range order.Items {
		var lot models.ProductLot
		if err := s.db.Where("id = ?", item.ProductLotID).First(&lot).Error; err != nil {
			return nil, err
		}

		previousQuantity := lot.CurrentQuantity
		lot.CurrentQuantity -= item.Quantity

		if err := s.db.Save(&lot).Error; err != nil {
			return nil, err
		}

		// Registrar movimentação de estoque
		movement := &models.StockMovement{
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
// Busca um pedido pelo ID
func (s *OrderService) GetByID(id uuid.UUID) (*OrderResponse, error) {
	var order models.Order
	if err := s.db.Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription").Where("id = ?", id).First(&order).Error; err != nil {
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
// Lista pedidos com filtros e paginação
func (s *OrderService) List(req ListOrderRequest) ([]OrderResponse, int64, error) {
	// --- 1. Definir paginação ---
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	// --- 2. Construir query ---
	query := s.db.Model(&models.Order{}).Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
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

	// --- 3. Contar total ---
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// --- 4. Buscar com paginação ---
	var orders []models.Order
	if err := query.Offset(offset).Limit(req.Limit).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	// --- 5. Converter para resposta ---
	var responses []OrderResponse
	for _, order := range orders {
		responses = append(responses, *s.toOrderResponse(&order))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATESTATUS()
// ================================================================
// Atualiza o status de um pedido
func (s *OrderService) UpdateStatus(id uuid.UUID, req UpdateOrderStatusRequest, userID uuid.UUID) (*OrderResponse, error) {
	// --- 1. Buscar pedido ---
	var order models.Order
	if err := s.db.Where("id = ?", id).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("pedido não encontrado")
		}
		return nil, err
	}

	// --- 2. Validar transição de status ---
	if !s.isValidStatusTransition(order.Status, req.Status) {
		return nil, fmt.Errorf("transição de status inválida: %s → %s", order.Status, req.Status)
	}

	// --- 3. Atualizar status ---
	order.Status = req.Status
	order.StatusUpdatedAt = time.Now()
	if req.Notes != "" {
		order.Notes = req.Notes
	}

	if err := s.db.Save(&order).Error; err != nil {
		return nil, err
	}

	// --- 4. Carregar relacionamentos ---
	if err := s.db.Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription").Where("id = ?", order.ID).First(&order).Error; err != nil {
		return nil, err
	}

	return s.toOrderResponse(&order), nil
}

// ================================================================
// FUNÇÃO UPDATETRACKING()
// ================================================================
// Adiciona código de rastreio ao pedido
func (s *OrderService) UpdateTracking(id uuid.UUID, req UpdateTrackingRequest) (*OrderResponse, error) {
	// --- 1. Buscar pedido ---
	var order models.Order
	if err := s.db.Where("id = ?", id).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("pedido não encontrado")
		}
		return nil, err
	}

	// --- 2. Verificar se o pedido está em status correto ---
	if order.Status != "correio" && order.Status != "entregue" {
		return nil, errors.New("pedido deve estar com status 'correio' ou 'entregue' para adicionar rastreio")
	}

	// --- 3. Atualizar rastreio ---
	order.TrackingCode = req.TrackingCode
	order.ShippingCarrier = req.ShippingCarrier

	if err := s.db.Save(&order).Error; err != nil {
		return nil, err
	}

	// --- 4. Carregar relacionamentos ---
	if err := s.db.Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription").Where("id = ?", order.ID).First(&order).Error; err != nil {
		return nil, err
	}

	return s.toOrderResponse(&order), nil
}

// ================================================================
// FUNÇÃO GETBYPATIENT()
// ================================================================
// Lista pedidos de um paciente específico
func (s *OrderService) GetByPatient(patientID uuid.UUID) ([]OrderResponse, error) {
	var orders []models.Order
	if err := s.db.Preload("Items").Preload("Items.ProductLot").Preload("Items.ProductLot.Product").
		Preload("Patient").Preload("Prescription").
		Where("patient_id = ?", patientID).Order("created_at DESC").
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
// Gera uma etiqueta de envio (simulado - retorna URL)
func (s *OrderService) GenerateLabel(id uuid.UUID) (string, error) {
	// --- 1. Verificar se o pedido existe ---
	var order models.Order
	if err := s.db.Where("id = ?", id).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("pedido não encontrado")
		}
		return "", err
	}

	// --- 2. Verificar se o pedido está em status correto ---
	if order.Status != "dispensa" && order.Status != "correio" {
		return "", errors.New("pedido deve estar com status 'dispensa' ou 'correio' para gerar etiqueta")
	}

	// --- 3. Gerar URL da etiqueta (simulado) ---
	labelURL := fmt.Sprintf("/api/orders/%s/label.pdf", order.ID.String())

	// --- 4. Salvar URL no pedido ---
	order.ShippingLabelURL = labelURL
	if err := s.db.Save(&order).Error; err != nil {
		return "", err
	}

	return labelURL, nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// isValidStatusTransition - Valida transição de status
func (s *OrderService) isValidStatusTransition(currentStatus, newStatus string) bool {
	// Mapa de transições válidas
	validTransitions := map[string][]string{
		"pendente":   {"separado", "cancelado"},
		"separado":   {"dispensa", "cancelado"},
		"dispensa":   {"correio", "cancelado"},
		"correio":    {"entregue", "cancelado"},
		"entregue":   {},
		"cancelado":  {},
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

// toOrderResponse - Converte models.Order para OrderResponse
func (s *OrderService) toOrderResponse(order *models.Order) *OrderResponse {
	items := []OrderItemResponse{}
	for _, item := range order.Items {
		productName := ""
		lotNumber := ""
		if item.ProductLot.ID != uuid.Nil {
			lotNumber = item.ProductLot.LotNumber
			if item.ProductLot.Product.ID != uuid.Nil {
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
	if order.Patient.ID != uuid.Nil {
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