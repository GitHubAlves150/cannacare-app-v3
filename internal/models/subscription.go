// ================================================================
// MODEL SUBSCRIPTION (ANUIDADES)
// ================================================================
// Gerencia as anuidades pagas pelos associados
//
// Tabela: subscriptions
//
// RELACIONAMENTOS:
//   - Pertence a um paciente (Patient)
//   - Tem um pagamento associado (Payment)
//
// STATUS:
//   pendente - Aguardando pagamento
//   pago     - Anuidade paga
//   atrasado - Vencida e não paga
//   cancelado - Cancelada
//
// REGRAS DE NEGÓCIO:
//   1. A anuidade é anual (validade de 1 ano)
//   2. Pacientes sociais podem ter isenção
//   3. Se estiver atrasado, o paciente não pode fazer pedidos
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
)



// Subscription representa uma anuidade do associado
type Subscription struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// PatientID - Paciente associado
	PatientID uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`

	// PaymentID - Pagamento relacionado (se pago)
	PaymentID *uuid.UUID `gorm:"type:uuid" json:"payment_id,omitempty"`

	// === DADOS DA ANUIDADE ===
	// DueDate - Data de vencimento
	DueDate time.Time `gorm:"not null" json:"due_date"`

	// Amount - Valor da anuidade
	Amount float64 `gorm:"type:decimal(10,2);not null" json:"amount"`

	// Status - Situação da anuidade
	Status string `gorm:"default:'pendente'" json:"status"`

	// PaidAt - Data/hora do pagamento
	PaidAt *time.Time `json:"paid_at,omitempty"`

	// === RELACIONAMENTOS ===
	Patient *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Payment *Payment `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
}

// TableName define o nome da tabela
func (Subscription) TableName() string {
	return "subscriptions"
}