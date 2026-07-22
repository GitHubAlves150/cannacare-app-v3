// ================================================================
// MODEL PATIENT_DOCUMENT (DOCUMENTOS DO PACIENTE)
// ================================================================
// Gerencia os documentos enviados pelos pacientes
//
// Tabela: patient_documents
//
// TIPOS DE DOCUMENTO:
//   rg_cpf               - RG ou CPF
//   comprovante_residencia - Comprovante de residência
//   laudo_medico         - Laudo médico
//   receita_medica       - Receita médica
//   autorizacao_anvisa   - Autorização da ANVISA
//   termo_consentimento  - Termo de consentimento
//
// STATUS DO DOCUMENTO:
//   em_analise - Aguardando análise
//   aprovado   - Documento aprovado
//   rejeitado  - Documento rejeitado
// ================================================================

package models

import (
	"github.com/google/uuid"
)


type PatientDocument struct {
	BaseModel

	PatientID    uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`
	ReviewedBy   *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`

	DocumentType string `gorm:"not null" json:"document_type"`
	FileURL      string `gorm:"not null" json:"file_url"`
	FileName     string `json:"file_name,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
	Status       string `gorm:"default:'em_analise'" json:"status"`
	ReviewedAt   *string `json:"reviewed_at,omitempty"`

	Patient  *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Reviewer *User    `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

func (PatientDocument) TableName() string {
	return "patient_documents"
}







/*
// PatientDocument representa um documento do paciente
type PatientDocument struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// PatientID - Paciente dono do documento
	PatientID uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`

	// ReviewedBy - Usuário que revisou o documento (se foi revisado)
	ReviewedBy *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`

	// === DADOS DO DOCUMENTO ===
	// DocumentType - Tipo do documento
	DocumentType string `gorm:"not null" json:"document_type"`

	// FileURL - Link/URL do arquivo (armazenado no S3 ou sistema de arquivos)
	FileURL string `gorm:"not null" json:"file_url"`

	// FileName - Nome original do arquivo
	FileName string `json:"file_name,omitempty"`

	// FileSize - Tamanho do arquivo em bytes
	FileSize int64 `json:"file_size,omitempty"`

	// MimeType - Tipo do arquivo (application/pdf, image/jpeg, etc)
	MimeType string `json:"mime_type,omitempty"`

	// Status - Situação do documento
	Status string `gorm:"default:'em_analise'" json:"status"`

	// ReviewedAt - Data/hora da revisão
	ReviewedAt *string `json:"reviewed_at,omitempty"`

	// === RELACIONAMENTOS ===
	Patient    *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Reviewer   *User    `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

// TableName define o nome da tabela
func (PatientDocument) TableName() string {
	return "patient_documents"
}

*/