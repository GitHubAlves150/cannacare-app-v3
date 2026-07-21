// ================================================================
// MODEL PRESCRIPTION (RECEITA MÉDICA)
// ================================================================
// Representa uma receita médica emitida para um paciente
//
// Tabela: prescriptions
//
// RELACIONAMENTOS:
//   - Pertence a um paciente (Patient)
//   - Pertence a um médico (Doctor)
//   - Tem muitos itens (PrescriptionItem)
//   - Tem muitos pedidos (Order)
//
// VALIDADE DA RECEITA:
//   - Data de emissão (issue_date) - Quando foi emitida
//   - Data de validade (expiration_date) - Até quando é válida
//   - Status: valida, proxima_vencer, vencida
//
// ALERTAS AUTOMÁTICOS:
//   - Próxima a vencer: 15 dias antes da expiration_date
//   - Vencida: bloqueia novos pedidos
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

	// === CHAVES ESTRANGEIRAS ===
	// PatientID - Paciente que recebeu a receita
	PatientID uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`

	// DoctorID - Médico que emitiu a receita
	DoctorID uuid.UUID `gorm:"type:uuid;not null" json:"doctor_id"`

	// === DADOS DA RECEITA ===
	// CID - Classificação Internacional de Doenças
	// Ex: G40.0 (Epilepsia), F32 (Depressão), etc
	CID string `gorm:"column:cid;not null" json:"cid"`  // ← COLUNA "cid"
	
	// IssueDate - Data de emissão da receita
	IssueDate time.Time `gorm:"not null" json:"issue_date"`

	// ExpirationDate - Data de validade da receita
	// Geralmente 1 ano após a emissão
	ExpirationDate time.Time `gorm:"not null" json:"expiration_date"`

	// Status - Situação atual da receita
	// valida, proxima_vencer, vencida
	Status string `gorm:"default:'valida'" json:"status"`

	// IsActive - Se a receita ainda está ativa
	IsActive bool `gorm:"default:true" json:"is_active"`

	// === ARQUIVO DA RECEITA ===
	// PrescriptionFileURL - Link para o arquivo PDF/JPG da receita
	PrescriptionFileURL string `json:"prescription_file_url,omitempty"`

	// PrescriptionFileName - Nome original do arquivo
	PrescriptionFileName string `json:"prescription_file_name,omitempty"`

	// === AUDITORIA ===
	// ValidatedBy - Quem validou a receita (se foi validada)
	ValidatedBy *uuid.UUID `gorm:"type:uuid" json:"validated_by,omitempty"`

	// ValidatedAt - Quando foi validada
	ValidatedAt *time.Time `json:"validated_at,omitempty"`

	// === RELACIONAMENTOS ===
	Patient *Patient          `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Doctor  *Doctor           `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Items   []PrescriptionItem `gorm:"foreignKey:PrescriptionID" json:"items,omitempty"`
	Orders  []Order           `gorm:"foreignKey:PrescriptionID" json:"orders,omitempty"`
}

// TableName define o nome da tabela
func (Prescription) TableName() string {
	return "prescriptions"
}