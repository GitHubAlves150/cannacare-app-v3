// ================================================================
// MODEL PATIENT_STATUS_HISTORY (HISTÓRICO DE STATUS)
// ================================================================
// Histórico completo de mudanças de status dos pacientes
//
// Tabela: patient_status_history
//
// POR QUE USAR?
// 1. Auditoria - Saber quem mudou o status e quando
// 2. Rastreabilidade - Acompanhar a jornada do paciente
// 3. Relatórios - Analisar tempo em cada status
// 4. Compliance - Registro para órgãos reguladores
//
// EXEMPLO:
//   full_name: "João Silva"
//   old_status: "pendente_documentacao"
//   new_status: "em_analise"
//   changed_by: "Maria (secretaria)"
//   created_at: "2026-07-21 10:00:00"
//   reason: "Documentos enviados e completos"
// ================================================================

package models

import (
	"github.com/google/uuid"
)

// PatientStatusHistory representa uma mudança de status de um paciente
type PatientStatusHistory struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// PatientID - Paciente que teve o status alterado
	PatientID uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`

	// ChangedBy - Usuário que fez a alteração
	ChangedBy *uuid.UUID `gorm:"type:uuid" json:"changed_by,omitempty"`

	// === STATUS ===
	// OldStatus - Status anterior
	OldStatus string `json:"old_status,omitempty"`

	// NewStatus - Novo status
	NewStatus string `gorm:"not null" json:"new_status"`

	// Reason - Motivo da mudança de status
	// Ex: "Documentos aprovados", "Paciente social", "Documentos rejeitados"
	Reason string `json:"reason,omitempty"`

	// === RELACIONAMENTOS ===
	Patient *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	ChangedByUser *User `gorm:"foreignKey:ChangedBy" json:"changed_by_user,omitempty"`
}

// TableName define o nome da tabela
func (PatientStatusHistory) TableName() string {
	return "patient_status_history"
}