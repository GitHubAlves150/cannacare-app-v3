// ================================================================
// PACOTE SERVICES - FINANCIAL SERVICE
// ================================================================
// Camada de serviço responsável pela gestão financeira.
//
// RESPONSABILIDADES:
// 1. Gerenciar anuidades dos associados
// 2. Registrar pagamentos (anuidade e produtos)
// 3. Emitir recibos/comprovantes
// 4. Consultar status financeiro do paciente
// 5. Relatórios de inadimplência
// 6. Alertas de vencimento
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
// STRUCT FINANCIALSERVICE
// ================================================================
type FinancialService struct {
	db *gorm.DB
}

// ================================================================
// FUNÇÃO NEWFINANCIALSERVICE()
// ================================================================
func NewFinancialService(db *gorm.DB) *FinancialService {
	return &FinancialService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

// CreateSubscriptionRequest - Dados para criar uma anuidade
type CreateSubscriptionRequest struct {
	PatientID string    `json:"patient_id" validate:"required"`
	DueDate   time.Time `json:"due_date" validate:"required"`
	Amount    float64   `json:"amount" validate:"required,min=0"`
}

// CreatePaymentRequest - Dados para registrar um pagamento
type CreatePaymentRequest struct {
	PatientID      string  `json:"patient_id" validate:"required"`
	OrderID        string  `json:"order_id" validate:"omitempty"`
	SubscriptionID string  `json:"subscription_id" validate:"omitempty"`
	PaymentType    string  `json:"payment_type" validate:"required,oneof=anuidade compra_produto doacao"`
	PaymentMethod  string  `json:"payment_method" validate:"required,oneof=pix boleto cartao transferencia"`
	Amount         float64 `json:"amount" validate:"required,min=0"`
	Installments   int     `json:"installments" validate:"omitempty,min=1"`
	ReceiptNumber  string  `json:"receipt_number" validate:"omitempty"`
	Status         string  `json:"status" validate:"omitempty,oneof=pendente pago recusado estornado"` // 🆕 ADICIONAR!
}

// UpdatePaymentStatusRequest - Para atualizar status do pagamento
type UpdatePaymentStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pago recusado estornado"`
}

// SubscriptionResponse - Resposta com dados da anuidade
type SubscriptionResponse struct {
	ID          string  `json:"id"`
	PatientID   string  `json:"patient_id"`
	PatientName string  `json:"patient_name"`
	DueDate     string  `json:"due_date"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	PaidAt      string  `json:"paid_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// PaymentResponse - Resposta com dados do pagamento
type PaymentResponse struct {
	ID             string  `json:"id"`
	PatientID      string  `json:"patient_id"`
	PatientName    string  `json:"patient_name"`
	OrderID        string  `json:"order_id,omitempty"`
	SubscriptionID string  `json:"subscription_id,omitempty"`
	PaymentType    string  `json:"payment_type"`
	PaymentMethod  string  `json:"payment_method"`
	Amount         float64 `json:"amount"`
	Installments   int     `json:"installments"`
	Status         string  `json:"status"`
	PaymentDate    string  `json:"payment_date,omitempty"`
	PaidAt         string  `json:"paid_at,omitempty"`
	ReceiptURL     string  `json:"receipt_url,omitempty"`
	ReceiptNumber  string  `json:"receipt_number,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// PatientFinancialStatus - Status financeiro do paciente
type PatientFinancialStatus struct {
	PatientID             string  `json:"patient_id"`
	PatientName           string  `json:"patient_name"`
	HasActiveSubscription bool    `json:"has_active_subscription"`
	SubscriptionStatus    string  `json:"subscription_status"`
	SubscriptionDueDate   string  `json:"subscription_due_date,omitempty"`
	TotalPaid             float64 `json:"total_paid"`
	TotalPending          float64 `json:"total_pending"`
}

// ListSubscriptionRequest - Filtros para listagem de anuidades
type ListSubscriptionRequest struct {
	PatientID string `json:"patient_id" query:"patient_id"`
	Status    string `json:"status" query:"status"`
	Page      int    `json:"page" query:"page"`
	Limit     int    `json:"limit" query:"limit"`
}

// ListPaymentRequest - Filtros para listagem de pagamentos
type ListPaymentRequest struct {
	PatientID   string `json:"patient_id" query:"patient_id"`
	PaymentType string `json:"payment_type" query:"payment_type"`
	Status      string `json:"status" query:"status"`
	Page        int    `json:"page" query:"page"`
	Limit       int    `json:"limit" query:"limit"`
}

// ================================================================
// FUNÇÃO CREATESUBSCRIPTION()
// ================================================================
// Cria uma nova anuidade para um paciente
func (s *FinancialService) CreateSubscription(req CreateSubscriptionRequest) (*SubscriptionResponse, error) {
	// --- 1. Validar paciente ---
	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		return nil, fmt.Errorf("ID do paciente inválido: %w", err)
	}

	var patient models.Patient
	if err := s.db.Where("id = ?", patientID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}

	// --- 2. Verificar se paciente já tem anuidade ativa ---
	var existingSubscription models.Subscription
	err = s.db.Where("patient_id = ? AND status IN (?)", patientID, []string{"pendente", "pago"}).
		First(&existingSubscription).Error
	if err == nil {
		return nil, errors.New("paciente já possui uma anuidade pendente ou paga")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- 3. Definir status da anuidade ---
	status := "pendente"
	if req.DueDate.Before(time.Now()) {
		status = "atrasado"
	}

	// --- 4. Criar anuidade ---
	subscription := &models.Subscription{
		PatientID: patientID,
		DueDate:   req.DueDate,
		Amount:    req.Amount,
		Status:    status,
	}

	if err := s.db.Create(subscription).Error; err != nil {
		return nil, err
	}

	return s.toSubscriptionResponse(subscription, &patient), nil
}

// ================================================================
// FUNÇÃO CREATEPAYMENT()
// ================================================================
// Registra um novo pagamento
func (s *FinancialService) CreatePayment(req CreatePaymentRequest) (*PaymentResponse, error) {
	// --- 1. Validar paciente ---
	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		return nil, fmt.Errorf("ID do paciente inválido: %w", err)
	}

	var patient models.Patient
	if err := s.db.Where("id = ?", patientID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}

	// --- 2. Validar payment type ---
	if req.PaymentType == "compra_produto" && req.OrderID == "" {
		return nil, errors.New("para compra de produto, order_id é obrigatório")
	}

	if req.PaymentType == "anuidade" && req.SubscriptionID == "" {
		return nil, errors.New("para anuidade, subscription_id é obrigatório")
	}

	// --- 3. Validar order_id se fornecido ---
	var orderID *uuid.UUID
	if req.OrderID != "" {
		id, err := uuid.Parse(req.OrderID)
		if err != nil {
			return nil, fmt.Errorf("ID do pedido inválido: %w", err)
		}
		orderID = &id

		// Verificar se o pedido existe
		var order models.Order
		if err := s.db.Where("id = ?", id).First(&order).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.New("pedido não encontrado")
			}
			return nil, err
		}
	}

	// --- 4. Validar subscription_id se fornecido ---
	var subscriptionID *uuid.UUID
	if req.SubscriptionID != "" {
		id, err := uuid.Parse(req.SubscriptionID)
		if err != nil {
			return nil, fmt.Errorf("ID da anuidade inválido: %w", err)
		}
		subscriptionID = &id

		// Verificar se a anuidade existe
		var subscription models.Subscription
		if err := s.db.Where("id = ?", id).First(&subscription).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.New("anuidade não encontrada")
			}
			return nil, err
		}
	}

	// --- 5. Criar pagamento ---
	payment := &models.Payment{
		PatientID:      patientID,
		OrderID:        orderID,
		SubscriptionID: subscriptionID,
		PaymentType:    req.PaymentType,
		PaymentMethod:  req.PaymentMethod,
		Amount:         req.Amount,
		Installments:   req.Installments,
		Status:         "pendente",
		ReceiptNumber:  req.ReceiptNumber,
	}

	if err := s.db.Create(payment).Error; err != nil {
		return nil, err
	}

	// --- 6. Se o pagamento for "pago", atualizar status ---
	if req.Status != "" && req.Status == "pago" {
		payment.Status = "pago"
		now := time.Now()
		payment.PaidAt = &now
		s.db.Save(payment)

		// Atualizar status da anuidade se for anuidade
		if req.PaymentType == "anuidade" && subscriptionID != nil {
			s.db.Model(&models.Subscription{}).
				Where("id = ?", subscriptionID).
				Updates(map[string]interface{}{
					"status":     "pago",
					"paid_at":    now,
					"payment_id": payment.ID,
				})
		}
	}

	return s.toPaymentResponse(payment, &patient), nil
}

// ================================================================
// FUNÇÃO GETSUBSCRIPTIONBYID()
// ================================================================
// Busca uma anuidade pelo ID
func (s *FinancialService) GetSubscriptionByID(id uuid.UUID) (*SubscriptionResponse, error) {
	var subscription models.Subscription
	if err := s.db.Preload("Patient").Where("id = ?", id).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("anuidade não encontrada")
		}
		return nil, err
	}
	return s.toSubscriptionResponse(&subscription, subscription.Patient), nil
}

// ================================================================
// FUNÇÃO LISTSUBSCRIPTIONS()
// ================================================================
// Lista anuidades com filtros
func (s *FinancialService) ListSubscriptions(req ListSubscriptionRequest) ([]SubscriptionResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	query := s.db.Model(&models.Subscription{}).Preload("Patient")

	if req.PatientID != "" {
		patientID, err := uuid.Parse(req.PatientID)
		if err == nil {
			query = query.Where("patient_id = ?", patientID)
		}
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var subscriptions []models.Subscription
	if err := query.Offset(offset).Limit(req.Limit).Order("due_date ASC").Find(&subscriptions).Error; err != nil {
		return nil, 0, err
	}

	var responses []SubscriptionResponse
	for _, sub := range subscriptions {
		responses = append(responses, *s.toSubscriptionResponse(&sub, sub.Patient))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO GETPAYMENTBYID()
// ================================================================
// Busca um pagamento pelo ID
func (s *FinancialService) GetPaymentByID(id uuid.UUID) (*PaymentResponse, error) {
	var payment models.Payment
	if err := s.db.Preload("Patient").Where("id = ?", id).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("pagamento não encontrado")
		}
		return nil, err
	}
	return s.toPaymentResponse(&payment, payment.Patient), nil
}

// ================================================================
// FUNÇÃO LISTPAYMENTS()
// ================================================================
// Lista pagamentos com filtros
func (s *FinancialService) ListPayments(req ListPaymentRequest) ([]PaymentResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	query := s.db.Model(&models.Payment{}).Preload("Patient")

	if req.PatientID != "" {
		patientID, err := uuid.Parse(req.PatientID)
		if err == nil {
			query = query.Where("patient_id = ?", patientID)
		}
	}
	if req.PaymentType != "" {
		query = query.Where("payment_type = ?", req.PaymentType)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var payments []models.Payment
	if err := query.Offset(offset).Limit(req.Limit).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	var responses []PaymentResponse
	for _, payment := range payments {
		responses = append(responses, *s.toPaymentResponse(&payment, payment.Patient))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATEPAYMENTSTATUS()
// ================================================================
// Atualiza o status de um pagamento
func (s *FinancialService) UpdatePaymentStatus(id uuid.UUID, req UpdatePaymentStatusRequest) (*PaymentResponse, error) {
	var payment models.Payment
	if err := s.db.Where("id = ?", id).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("pagamento não encontrado")
		}
		return nil, err
	}

	// Atualizar status
	payment.Status = req.Status
	if req.Status == "pago" {
		now := time.Now()
		payment.PaidAt = &now
	}

	if err := s.db.Save(&payment).Error; err != nil {
		return nil, err
	}

	// Se for pagamento de anuidade e foi pago, atualizar anuidade
	if payment.PaymentType == "anuidade" && payment.SubscriptionID != nil && req.Status == "pago" {
		s.db.Model(&models.Subscription{}).
			Where("id = ?", payment.SubscriptionID).
			Updates(map[string]interface{}{
				"status":     "pago",
				"paid_at":    payment.PaidAt,
				"payment_id": payment.ID,
			})
	}

	var patient models.Patient
	s.db.Where("id = ?", payment.PatientID).First(&patient)

	return s.toPaymentResponse(&payment, &patient), nil
}

// ================================================================
// FUNÇÃO GETPATIENTFINANCIALSTATUS()
// ================================================================
// Retorna o status financeiro de um paciente
func (s *FinancialService) GetPatientFinancialStatus(patientID uuid.UUID) (*PatientFinancialStatus, error) {
	var patient models.Patient
	if err := s.db.Where("id = ?", patientID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}

	// Buscar anuidade ativa
	var subscription models.Subscription
	hasActive := false
	subStatus := "nenhuma"
	subDueDate := ""

	err := s.db.Where("patient_id = ? AND status IN (?)", patientID, []string{"pendente", "pago", "atrasado"}).
		Order("due_date DESC").First(&subscription).Error
	if err == nil {
		hasActive = true
		subStatus = subscription.Status
		subDueDate = subscription.DueDate.Format("2006-01-02")
	}

	// Calcular totais pagos e pendentes
	var totalPaid float64
	s.db.Model(&models.Payment{}).
		Where("patient_id = ? AND status = ?", patientID, "pago").
		Select("COALESCE(SUM(amount), 0)").Scan(&totalPaid)

	var totalPending float64
	s.db.Model(&models.Payment{}).
		Where("patient_id = ? AND status IN (?)", patientID, []string{"pendente", "recusado"}).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalPending)

	return &PatientFinancialStatus{
		PatientID:             patient.ID.String(),
		PatientName:           patient.FullName,
		HasActiveSubscription: hasActive,
		SubscriptionStatus:    subStatus,
		SubscriptionDueDate:   subDueDate,
		TotalPaid:             totalPaid,
		TotalPending:          totalPending,
	}, nil
}

// ================================================================
// FUNÇÃO GETOVERDUESUBSCRIPTIONS()
// ================================================================
// Retorna anuidades em atraso (usando view)
func (s *FinancialService) GetOverdueSubscriptions() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_overdue_subscriptions").Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// toSubscriptionResponse - Converte models.Subscription para SubscriptionResponse
func (s *FinancialService) toSubscriptionResponse(sub *models.Subscription, patient *models.Patient) *SubscriptionResponse {
	patientName := ""
	if patient != nil && patient.ID != uuid.Nil {
		patientName = patient.FullName
	}

	paidAt := ""
	if sub.PaidAt != nil {
		paidAt = sub.PaidAt.Format("2006-01-02 15:04:05")
	}

	return &SubscriptionResponse{
		ID:          sub.ID.String(),
		PatientID:   sub.PatientID.String(),
		PatientName: patientName,
		DueDate:     sub.DueDate.Format("2006-01-02"),
		Amount:      sub.Amount,
		Status:      sub.Status,
		PaidAt:      paidAt,
		CreatedAt:   sub.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   sub.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toPaymentResponse - Converte models.Payment para PaymentResponse
func (s *FinancialService) toPaymentResponse(payment *models.Payment, patient *models.Patient) *PaymentResponse {
	patientName := ""
	if patient != nil && patient.ID != uuid.Nil {
		patientName = patient.FullName
	}

	orderID := ""
	if payment.OrderID != nil {
		orderID = payment.OrderID.String()
	}

	subscriptionID := ""
	if payment.SubscriptionID != nil {
		subscriptionID = payment.SubscriptionID.String()
	}

	paymentDate := ""
	if payment.PaymentDate != nil {
		paymentDate = payment.PaymentDate.Format("2006-01-02")
	}

	paidAt := ""
	if payment.PaidAt != nil {
		paidAt = payment.PaidAt.Format("2006-01-02 15:04:05")
	}

	return &PaymentResponse{
		ID:             payment.ID.String(),
		PatientID:      payment.PatientID.String(),
		PatientName:    patientName,
		OrderID:        orderID,
		SubscriptionID: subscriptionID,
		PaymentType:    payment.PaymentType,
		PaymentMethod:  payment.PaymentMethod,
		Amount:         payment.Amount,
		Installments:   payment.Installments,
		Status:         payment.Status,
		PaymentDate:    paymentDate,
		PaidAt:         paidAt,
		ReceiptURL:     payment.ReceiptURL,
		ReceiptNumber:  payment.ReceiptNumber,
		CreatedAt:      payment.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      payment.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
