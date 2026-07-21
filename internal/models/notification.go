// ================================================================
// MODEL NOTIFICATION (NOTIFICAÇÕES)
// ================================================================
// Sistema de notificações e alertas para usuários
//
// Tabela: notifications
//
// TIPOS DE NOTIFICAÇÃO:
//   prescription_expiring - Receita próxima a vencer (15 dias)
//   prescription_expired  - Receita vencida
//   low_stock             - Estoque baixo
//   product_expiring      - Produto com validade próxima
//   payment_due           - Pagamento próximo ao vencimento
//   order_status          - Mudança de status do pedido
//
// RELACIONAMENTOS:
//   - Pertence a um usuário (User) - quem recebe a notificação
//   - Pode estar associado a um paciente (Patient)
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification representa uma notificação para um usuário
type Notification struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// UserID - Usuário que recebe a notificação
	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`

	// PatientID - Paciente relacionado (se aplicável)
	PatientID *uuid.UUID `gorm:"type:uuid" json:"patient_id,omitempty"`

	// === DADOS DA NOTIFICAÇÃO ===
	// Type - Tipo da notificação
	Type string `gorm:"not null" json:"type"`

	// Title - Título da notificação
	Title string `gorm:"not null" json:"title"`

	// Message - Mensagem detalhada
	Message string `gorm:"not null" json:"message"`

	// ReadAt - Data/hora em que foi lida
	// Se null, ainda não foi lida
	ReadAt *time.Time `json:"read_at,omitempty"`

	// ActionURL - URL para ação (ex: link para o pedido)
	ActionURL string `json:"action_url,omitempty"`

	// === RELACIONAMENTOS ===
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Patient *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
}

// TableName define o nome da tabela
func (Notification) TableName() string {
	return "notifications"
}
