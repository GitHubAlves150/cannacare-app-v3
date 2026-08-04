// ================================================================
// CANNACARE - MODEL PRESCRIPTION (RECEITA MÉDICA)
// ================================================================
// Representa uma receita médica emitida para um paciente
//
// TABELA: prescriptions
//
// RELACIONAMENTOS:
//   - Cada prescrição pertence a UMA associação (association_id)
//   - Pertence a um paciente (Patient)
//   - Pertence a um médico (Doctor)
//   - Tem muitos itens (PrescriptionItem)
//   - Tem muitos pedidos (Order)
//
// VALIDADE DA RECEITA:
//   - Data de emissão (issue_date) - Quando foi emitida
//   - Data de validade (expiration_date) - Até quando é válida
//   - Status: valida, proxima_vencer, vencida
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
)

// Prescription representa uma receita médica
type Prescription struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVE ESTRANGEIRA PARA MULTI-TENANCY ===
	// AssociationID - QUAL associação esta prescrição pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	// === CHAVES ESTRANGEIRAS ===
	PatientID uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`
	DoctorID  uuid.UUID `gorm:"type:uuid;not null" json:"doctor_id"`

	// === DADOS DA RECEITA ===
	// CID - Classificação Internacional de Doenças
	CID string `gorm:"column:cid;not null" json:"cid"`

	// IssueDate - Data de emissão da receita
	IssueDate time.Time `gorm:"not null" json:"issue_date"`

	// ExpirationDate - Data de validade da receita
	ExpirationDate time.Time `gorm:"not null" json:"expiration_date"`

	// Status - Situação atual da receita
	Status string `gorm:"default:'valida'" json:"status"`

	// IsActive - Se a receita ainda está ativa
	IsActive bool `gorm:"default:true" json:"is_active"`

	// === ARQUIVO DA RECEITA ===
	PrescriptionFileURL  string `json:"prescription_file_url,omitempty"`
	PrescriptionFileName string `json:"prescription_file_name,omitempty"`

	// === AUDITORIA ===
	ValidatedBy *uuid.UUID `gorm:"type:uuid" json:"validated_by,omitempty"`
	ValidatedAt *time.Time `json:"validated_at,omitempty"`

	// === RELACIONAMENTOS ===
	Association *Association       `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Patient     *Patient           `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Doctor      *Doctor            `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Items       []PrescriptionItem `gorm:"foreignKey:PrescriptionID" json:"items,omitempty"`
	Orders      []Order            `gorm:"foreignKey:PrescriptionID" json:"orders,omitempty"`
}

// TableName define o nome da tabela
func (Prescription) TableName() string {
	return "prescriptions"
}