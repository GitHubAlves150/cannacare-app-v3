// ================================================================
// CANNACARE - MODEL DOCTOR (MÉDICO PRESCRITOR)
// ================================================================
// Representa um médico que emite receitas para os pacientes
//
// TABELA: doctors
//
// RELACIONAMENTOS:
//   - Cada médico pertence a UMA associação (association_id)
//   - Um médico pode ter muitas prescrições (Prescription)
//
// REGRAS DE NEGÓCIO:
//   - O CRM é único DENTRO da associação
//   - A constraint UNIQUE(crm, crm_state, association_id) garante isso
//   - Médicos podem ser ativos/inativos (is_active)
// ================================================================

package models

import (
	"github.com/google/uuid"
)

// Doctor representa um médico prescritor
type Doctor struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVE ESTRANGEIRA PARA MULTI-TENANCY ===
	// AssociationID - QUAL associação este médico pertence
	// ⚠️ NUNCA deixar nulo! Todo médico pertence a uma associação
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	// === DADOS PESSOAIS ===
	// Name - Nome completo do médico
	Name string `gorm:"not null" json:"name"`

	// CRM - Número do registro no Conselho Regional de Medicina
	// Obrigatório e único (em combinação com CRMState e AssociationID)
	CRM string `gorm:"not null" json:"crm"`

	// CRMState - Estado onde o CRM foi registrado (ex: SP, RJ, MG)
	CRMState string `gorm:"not null" json:"crm_state"`

	// Specialty - Especialidade médica (Neurologia, Psiquiatria, etc)
	Specialty string `json:"specialty,omitempty"`

	// === CONTATO ===
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`

	// === STATUS ===
	// IsActive - Se o médico ainda está ativo na associação
	IsActive bool `gorm:"default:true" json:"is_active"`

	// === RELACIONAMENTOS ===
	Association  *Association  `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Prescriptions []Prescription `gorm:"foreignKey:DoctorID" json:"prescriptions,omitempty"`
}

// TableName define o nome da tabela
func (Doctor) TableName() string {
	return "doctors"
}