// ================================================================
// CANNACARE - MODEL ASSOCIATION (ASSOCIAÇÃO)
// ================================================================
// Este é o model da tabela MESTRA do sistema.
// Cada associação (cliente) tem um registro aqui.
// Todas as outras tabelas se VINCULAM a ela via association_id.
//
// TABELA: associations
//
// O QUE É UMA ASSOCIAÇÃO?
//   - É um cliente do SaaS (ex: "Associação Canábica do Brasil")
//   - Cada associação tem seus próprios pacientes, médicos, receitas, etc.
//   - Os dados de diferentes associações são ISOLADOS no banco
//
// CAMPOS IMPORTANTES:
//   - plan: basic (limitado), premium (ilimitado), enterprise (personalizado)
//   - status: pending (cadastrou mas não pagou), active (ativo), suspended (suspenso)
//   - patient_limit: número máximo de pacientes (ex: basic = 50, premium = 999999)
// ================================================================

package models

import (
	"time"
)

// Association representa uma associação (cliente do SaaS)
type Association struct {
	// === CAMPOS BASE (EMBUTIDOS) ===
	// Herda: ID, CreatedAt, UpdatedAt, DeletedAt
	BaseModel

	// === DADOS DA ASSOCIAÇÃO ===
	// Name - Nome da associação (ex: "Associação Canábica do Brasil")
	Name string `gorm:"not null" json:"name"`

	// CNPJ - Cadastro Nacional de Pessoa Jurídica (único)
	// É NOT NULL e UNIQUE, pois cada associação tem um CNPJ único
	CNPJ string `gorm:"unique;not null" json:"cnpj"`

	// Email - Email principal da associação
	// Também único (não pode ter duas associações com o mesmo email)
	Email string `gorm:"unique;not null" json:"email"`

	// Phone - Telefone de contato
	Phone string `json:"phone,omitempty"`

	// Address - Endereço da associação
	Address string `json:"address,omitempty"`

	// === PLANO E CONTROLE ===
	// Plan - Plano contratado: basic, premium, enterprise
	// basic = limitado, premium = ilimitado, enterprise = personalizado
	Plan string `gorm:"default:'basic';check:plan IN ('basic','premium','enterprise')" json:"plan"`

	// Status - Situação da associação
	// pending = aguardando pagamento, active = ativa, suspended = suspensa, cancelled = cancelada
	Status string `gorm:"default:'pending';check:status IN ('pending','active','suspended','cancelled')" json:"status"`

	// PatientLimit - Número máximo de pacientes permitidos
	// basic = 50, premium = 999999 (ilimitado), enterprise = personalizado
	PatientLimit int `gorm:"default:50" json:"patient_limit"`

	// TrialEndsAt - Data de término do período de teste (se houver)
	TrialEndsAt *time.Time `json:"trial_ends_at,omitempty"`

	// === INTEGRAÇÃO COM GATEWAY DE PAGAMENTO ===
	// StripeCustomerID - ID do cliente no Stripe (para cobranças)
	StripeCustomerID string `gorm:"type:varchar(100)" json:"stripe_customer_id,omitempty"`

	// SubscriptionID - ID da assinatura no Stripe
	SubscriptionID string `gorm:"type:varchar(100)" json:"subscription_id,omitempty"`
}

// TableName define o nome da tabela no PostgreSQL
// Sempre usar plural e minúsculo (boa prática)
func (Association) TableName() string {
	return "associations"
}
