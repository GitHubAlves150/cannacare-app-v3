// ================================================================
// MODEL DOCTOR
// ================================================================
// Representa um médico prescritor que emite receitas para os pacientes
//
// Tabela: doctors
//
// RELACIONAMENTOS:
//   - Um médico pode ter muitas prescrições (Prescription)
//
// CAMPOS IMPORTANTES:
//   - CRM (Conselho Regional de Medicina) - Obrigatório e único
//   - UF (Estado) - Onde o médico está registrado
//   - Especialidade - Neurologia, Psiquiatria, etc
// ================================================================

package models

// Doctor representa um médico prescritor
type Doctor struct {
	// === CAMPOS BASE ===
	BaseModel

	// === DADOS PESSOAIS ===
	// Name - Nome completo do médico
	Name string `gorm:"not null" json:"name"`

	// CRM - Número do registro no Conselho Regional de Medicina
	// Obrigatório e único (em combinação com CRMState)
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
	// Prescriptions - Lista de receitas emitidas por este médico
	Prescriptions []Prescription `gorm:"foreignKey:DoctorID" json:"prescriptions,omitempty"`
}

// TableName define o nome da tabela
func (Doctor) TableName() string {
	return "doctors"
}