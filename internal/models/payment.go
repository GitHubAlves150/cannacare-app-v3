// ================================================================
// MODEL PAYMENT (PAGAMENTOS)
// ================================================================
// Registro de todos os pagamentos financeiros
//
// Tabela: payments
//
// TIPOS DE PAGAMENTO:
//   anuidade        - Pagamento da anuidade do associado
//   compra_produto  - Pagamento de produtos (pedidos)
//   doacao          - Doação para a associação
//
// FORMAS DE PAGAMENTO:
//   pix, boleto, cartao, transferencia
//
// RELACIONAMENTOS:
//   - Pertence a um paciente (Patient)
//   - Pode estar associado a um pedido (Order) - se for compra de produto
//   - Pode estar associado a uma anuidade (Subscription) - se for anuidade
//
// STATUS:
//   pendente - Aguardando confirmação
//   pago     - Confirmado
//   recusado - Negado pelo gateway
//   estornado - Estornado
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Payment representa um pagamento financeiro
type Payment struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// PatientID - Paciente que fez o pagamento
	PatientID uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`

	// OrderID - Pedido relacionado (se for compra de produto)
	OrderID *uuid.UUID `gorm:"type:uuid" json:"order_id,omitempty"`

	// SubscriptionID - Anuidade relacionada (se for anuidade)
	SubscriptionID *uuid.UUID `gorm:"type:uuid" json:"subscription_id,omitempty"`

	// === DADOS DO PAGAMENTO ===
	// PaymentType - Tipo de pagamento: anuidade, compra_produto, doacao
	PaymentType string `gorm:"not null" json:"payment_type"`

	// PaymentMethod - Forma de pagamento: pix, boleto, cartao, transferencia
	PaymentMethod string `gorm:"not null" json:"payment_method"`

	// Amount - Valor do pagamento
	Amount float64 `gorm:"type:decimal(10,2);not null" json:"amount"`

	// Installments - Número de parcelas (1 = à vista)
	Installments int `gorm:"default:1" json:"installments"`

	// === STATUS ===
	// Status - Situação do pagamento
	Status string `gorm:"default:'pendente'" json:"status"`

	// PaymentDate - Data do pagamento (quando foi feito)
	PaymentDate *time.Time `json:"payment_date,omitempty"`

	// PaidAt - Data/hora da confirmação do pagamento
	PaidAt *time.Time `json:"paid_at,omitempty"`

	// === COMPROVANTES ===
	// ReceiptURL - Link para o comprovante/recibo (PDF ou imagem)
	ReceiptURL string `json:"receipt_url,omitempty"`

	// ReceiptNumber - Número do recibo/NF
	ReceiptNumber string `json:"receipt_number,omitempty"`

	// === INTEGRAÇÃO COM GATEWAY ===
	// TransactionID - ID da transação no gateway (Stripe, PagSeguro, etc)
	TransactionID string `json:"transaction_id,omitempty"`

	// GatewayResponse - Resposta completa do gateway (JSON)
	// Armazena dados brutos para debug e auditoria
	GatewayResponse map[string]interface{} `gorm:"type:jsonb" json:"gateway_response,omitempty"`

	// === RELACIONAMENTOS ===
	Patient      *Patient      `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Order        *Order        `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Subscription *Subscription `gorm:"foreignKey:SubscriptionID" json:"subscription,omitempty"`
}

// TableName define o nome da tabela
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate - Validação antes de criar o pagamento
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	// Se não houver status definido, usa 'pendente'
	if p.Status == "" {
		p.Status = "pendente"
	}
	return nil
}