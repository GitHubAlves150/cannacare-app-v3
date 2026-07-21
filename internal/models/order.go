// ================================================================
// MODEL ORDER (PEDIDOS)
// ================================================================
// Pedidos de medicação feitos pelos pacientes
//
// Tabela: orders
//
// STATUS DO PEDIDO:
//   pendente  - Aguardando processamento
//   separado  - Produto separado no estoque
//   dispensa  - Entregue para a farmácia/dispensa
//   correio   - Enviado pelo correio (com código de rastreio)
//   entregue  - Entregue ao paciente
//   cancelado - Pedido cancelado
//
// RELACIONAMENTOS:
//   - Pertence a um paciente (Patient)
//   - Pertence a uma receita (Prescription)
//   - Tem muitos itens (OrderItem)
//   - Tem um pagamento associado (Payment)
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
)

// Order representa um pedido de medicação
type Order struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// PatientID - Paciente que fez o pedido
	PatientID uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`

	// PrescriptionID - Receita usada para este pedido
	PrescriptionID uuid.UUID `gorm:"type:uuid;not null" json:"prescription_id"`

	// === STATUS E CONTROLE ===
	// Status - Situação atual do pedido
	Status string `gorm:"not null;default:'pendente'" json:"status"`

	// StatusUpdatedAt - Última atualização do status
	StatusUpdatedAt time.Time `gorm:"autoUpdateTime" json:"status_updated_at"`

	// === DADOS DO PEDIDO ===
	// TotalAmount - Valor total do pedido
	TotalAmount float64 `gorm:"type:decimal(10,2);not null" json:"total_amount"`

	// Notes - Observações do pedido
	Notes string `json:"notes,omitempty"`

	// OrderDate - Data em que o pedido foi feito
	OrderDate time.Time `gorm:"autoCreateTime" json:"order_date"`

	// === ENVIO E RASTREIO ===
	// ShippingCarrier - Transportadora (Correios, Jadlog, etc)
	ShippingCarrier string `json:"shipping_carrier,omitempty"`

	// TrackingCode - Código de rastreio (ex: "AB123456789BR")
	TrackingCode string `json:"tracking_code,omitempty"`

	// ShippingLabelURL - Link para a etiqueta de envio (PDF)
	ShippingLabelURL string `json:"shipping_label_url,omitempty"`

	// ShippingCost - Custo do frete
	ShippingCost float64 `gorm:"type:decimal(10,2);default:0" json:"shipping_cost"`

	// === RELACIONAMENTOS ===
	Patient      *Patient      `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Prescription *Prescription `gorm:"foreignKey:PrescriptionID" json:"prescription,omitempty"`
	Items        []OrderItem   `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Payment      *Payment      `gorm:"foreignKey:OrderID" json:"payment,omitempty"`
}

// TableName define o nome da tabela
func (Order) TableName() string {
	return "orders"
}
