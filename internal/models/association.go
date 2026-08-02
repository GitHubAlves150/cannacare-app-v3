// ================================================================
// CANNACARE - MODEL ASSOCIATION (ASSOCIAÇÃO)
// ================================================================
// ⚠️ Este arquivo SUBSTITUI a versão anterior — adiciona os campos
// de vigência do plano (usados pelo cadastro público do site e,
// depois, pelo job de expiração/renovação).
// ================================================================

package models

import (
	"time"
)

type Association struct {
	BaseModel

	// === DADOS DA ASSOCIAÇÃO ===
	Name    string `gorm:"not null" json:"name"`
	CNPJ    string `gorm:"unique;not null" json:"cnpj"`
	Email   string `gorm:"unique;not null" json:"email"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`

	// === PLANO E CONTROLE ===
	// Plan - basic, premium, enterprise
	Plan string `gorm:"default:'basic';check:plan IN ('basic','premium','enterprise')" json:"plan"`

	// Status - pending (aguardando pagamento), active, suspended, cancelled
	Status string `gorm:"default:'pending';check:status IN ('pending','active','suspended','cancelled')" json:"status"`

	PatientLimit int `gorm:"default:50" json:"patient_limit"`
	UserLimit    int `gorm:"default:3" json:"user_limit"`

	// === VIGÊNCIA DO PLANO ===
	// PlanActivatedAt - quando o plano atual (ou a última renovação) começou
	PlanActivatedAt *time.Time `json:"plan_activated_at,omitempty"`

	// PlanExpiresAt - quando o plano atual expira (nil = plano gratuito, sem prazo)
	PlanExpiresAt *time.Time `json:"plan_expires_at,omitempty"`

	// PaymentReference - ID da preferência/pagamento no Mercado Pago
	PaymentReference string `json:"payment_reference,omitempty"`

	TrialEndsAt *time.Time `json:"trial_ends_at,omitempty"`
}

func (Association) TableName() string {
	return "associations"
}