// ================================================================
// MODEL ANAMNESE (ACOLHIMENTO E ACOMPANHAMENTO)
// ================================================================
// Registro do acolhimento inicial e rastreamentos periódicos
//
// Tabela: anamneses
//
// TIPOS:
//   inicial            - Primeiro acolhimento do paciente
//   rastreio_1_mes     - Acompanhamento após 1 mês
//   rastreio_3_meses   - Acompanhamento após 3 meses
//   rastreio_6_meses   - Acompanhamento após 6 meses
//   acompanhamento_continuo - Acompanhamento contínuo
//
// RELACIONAMENTOS:
//   - Pertence a um paciente (Patient)
//   - Pertence a um usuário (User) - quem realizou o acolhimento
//
// CAMPOS ESTRUTURADOS PARA RELATÓRIOS:
//   - Sintomas e intensidade (1-10)
//   - Efeitos colaterais e intensidade
//   - Adesão ao tratamento (alta/média/baixa)
//   - Desafios e melhorias
//   - Dados clínicos (peso, pressão, frequência cardíaca)
//   - Extra em JSON (flexível para campos adicionais)
// ================================================================

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Anamnese representa um registro de acolhimento/acompanhamento
type Anamnese struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// PatientID - Paciente avaliado
	PatientID uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`

	// ResponsibleUserID - Profissional que realizou o acolhimento
	ResponsibleUserID uuid.UUID `gorm:"type:uuid;not null" json:"responsible_user_id"`

	// === TIPO DE ANAMNESE ===
	// Type - inicial, rastreio_1_mes, rastreio_3_meses, etc
	Type string `gorm:"not null" json:"type"`

	// === AVALIAÇÃO CLÍNICA ===
	// Symptoms - Descrição dos sintomas relatados
	Symptoms string `json:"symptoms,omitempty"`

	// SymptomIntensity - Intensidade dos sintomas (escala 1-10)
	SymptomIntensity *int `gorm:"check:symptom_intensity >= 1 AND symptom_intensity <= 10" json:"symptom_intensity,omitempty"`

	// SideEffects - Efeitos colaterais relatados
	SideEffects string `json:"side_effects,omitempty"`

	// SideEffectIntensity - Intensidade dos efeitos colaterais (1-10)
	SideEffectIntensity *int `gorm:"check:side_effect_intensity >= 1 AND side_effect_intensity <= 10" json:"side_effect_intensity,omitempty"`

	// === ADESÃO AO TRATAMENTO ===
	// TreatmentAdherence - Alta, Média, Baixa
	TreatmentAdherence string `gorm:"check:treatment_adherence IN ('alta', 'media', 'baixa')" json:"treatment_adherence,omitempty"`

	// Challenges - Desafios enfrentados no tratamento
	Challenges string `json:"challenges,omitempty"`

	// Improvements - Melhorias observadas
	Improvements string `json:"improvements,omitempty"`

	// AdditionalNotes - Observações adicionais
	AdditionalNotes string `json:"additional_notes,omitempty"`

	// === DADOS CLÍNICOS (OPCIONAIS) ===
	Weight        *float64 `gorm:"type:decimal(5,2)" json:"weight,omitempty"`
	BloodPressure string   `json:"blood_pressure,omitempty"`
	HeartRate     *int     `json:"heart_rate,omitempty"`

	// === DADOS FLEXÍVEIS ===
	// ExtraResponses - Campos adicionais em JSON
	// Permite adicionar campos sem alterar a estrutura da tabela
	ExtraResponses map[string]interface{} `gorm:"type:jsonb" json:"extra_responses,omitempty"`

	// === RELACIONAMENTOS ===
	Patient         *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	ResponsibleUser *User    `gorm:"foreignKey:ResponsibleUserID" json:"responsible_user,omitempty"`
}

// TableName define o nome da tabela
func (Anamnese) TableName() string {
	return "anamneses"
}

// BeforeCreate - Hook antes de criar o registro
func (a *Anamnese) BeforeCreate(tx *gorm.DB) error {
	// Se não houver tipo definido, usa 'inicial'
	if a.Type == "" {
		a.Type = "inicial"
	}
	return nil
}
