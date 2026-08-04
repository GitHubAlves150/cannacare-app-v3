// ================================================================
// CANNACARE - MODEL PATIENT (PACIENTE)
// ================================================================
// Representa um paciente/associado da associação
//
// TABELA: patients
//
// RELACIONAMENTOS:
//   - Cada paciente pertence a UMA associação (association_id)
//   - Um paciente pode ter um usuário (User) para acesso ao portal
//   - Um paciente pode ter muitos documentos (PatientDocument)
//   - Um paciente pode ter muitas receitas (Prescription)
//   - Um paciente pode ter muitos pedidos (Order)
//   - Um paciente pode ter muitas anamneses (Anamnese)
//   - Um paciente pode ter muitos pagamentos (Payment)
//   - Um paciente pode ter muitas anuidades (Subscription)
//
// STATUS DO PACIENTE:
//   pendente_documentacao - Aguardando envio de documentos
//   em_analise           - Documentos sendo analisados
//   aprovado             - Paciente aprovado e ativo
//   negado               - Paciente reprovado
//   assistente_social    - Em análise social (vulnerabilidade)
//
// REGRA DE NEGÓCIO CRÍTICA:
//   - Pacientes SÓ podem ser cadastrados se a associação não tiver
//     ultrapassado o patient_limit (definido no plano)
//   - Isso é garantido pela TRIGGER enforce_patient_limit_before_insert
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Patient representa um paciente da associação
type Patient struct {
	// === CAMPOS BASE (EMBUTIDOS) ===
	BaseModel

	// === CHAVE ESTRANGEIRA PARA MULTI-TENANCY ===
	// AssociationID - QUAL associação este paciente pertence
	// ⚠️ NUNCA deixar nulo! Todo paciente pertence a uma associação
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	// === RELACIONAMENTO COM USUÁRIO ===
	// UserID - Chave estrangeira para users (pode ser NULL se não tiver acesso ao portal)
	    // 🔧 UserID agora é opcional e pode ser NULL
    UserID *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"` // ← Remove "unique"

	// === DADOS PESSOAIS (OBRIGATÓRIOS) ===
	// FullName - Nome completo do paciente
	FullName string `gorm:"not null" json:"full_name"`

	// BirthDate - Data de nascimento (formato: YYYY-MM-DD)
	BirthDate time.Time `gorm:"not null" json:"birth_date"`

	// Gender - Gênero: Masculino, Feminino, Outro
	Gender string `json:"gender,omitempty"`

	// === DOCUMENTOS (OBRIGATÓRIOS) ===
	// CPF - Cadastro de Pessoa Física (único DENTRO da associação)
	// A constraint UNIQUE(cpf, association_id) garante isso
	CPF string `gorm:"not null" json:"cpf"`

	// RG - Registro Geral (opcional)
	RG string `json:"rg,omitempty"`

	// === CONTATO ===
	Phone    string `json:"phone,omitempty"`
	WhatsApp string `gorm:"column:whatsapp" json:"whatsapp"`
	Email    string `json:"email,omitempty"`

	// === ENDEREÇO ===
	AddressStreet       string `json:"address_street,omitempty"`
	AddressNumber       string `json:"address_number,omitempty"`
	AddressComplement   string `json:"address_complement,omitempty"`
	AddressNeighborhood string `json:"address_neighborhood,omitempty"`
	AddressCity         string `json:"address_city,omitempty"`
	AddressState        string `json:"address_state,omitempty"`
	AddressZipCode      string `json:"address_zipcode,omitempty"`

	// === STATUS E CONTROLE ===
	// Status - Situação atual do paciente no sistema
	Status string `gorm:"not null;default:'pendente_documentacao'" json:"status"`

	// IsSocialPatient - Indica se o paciente é social (vulnerável)
	// Pacientes sociais podem ter isenção de anuidade
	IsSocialPatient bool `gorm:"default:false" json:"is_social_patient"`

	// SocialAssistantNotes - Observações do assistente social
	SocialAssistantNotes string `json:"social_assistant_notes,omitempty"`

	// ApprovedAt - Data em que o paciente foi aprovado
	ApprovedAt *time.Time `json:"approved_at,omitempty"`

	// === RELACIONAMENTOS ===
	User          *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Association   *Association      `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Documents     []PatientDocument `gorm:"foreignKey:PatientID" json:"documents,omitempty"`
	Prescriptions []Prescription    `gorm:"foreignKey:PatientID" json:"prescriptions,omitempty"`
	Orders        []Order           `gorm:"foreignKey:PatientID" json:"orders,omitempty"`
	Anamneses     []Anamnese        `gorm:"foreignKey:PatientID" json:"anamneses,omitempty"`
	Subscriptions []Subscription    `gorm:"foreignKey:PatientID" json:"subscriptions,omitempty"`
	Payments      []Payment         `gorm:"foreignKey:PatientID" json:"payments,omitempty"`
}

// TableName define o nome da tabela
func (Patient) TableName() string {
	return "patients"
}

// BeforeCreate - Hook executado antes de criar o paciente
func (p *Patient) BeforeCreate(tx *gorm.DB) error {
	// Se o status não foi definido, usa o padrão
	if p.Status == "" {
		p.Status = "pendente_documentacao"
	}
	return nil
}