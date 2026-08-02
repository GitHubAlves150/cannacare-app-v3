// ================================================================
// MODEL ANAMNESE (ACOLHIMENTO E ACOMPANHAMENTO)
// ================================================================
// ⚠️ Adicionado AssociationID para multi-tenancy.
// ================================================================

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Anamnese struct {
	BaseModel

	// AssociationID - QUAL associação esta anamnese pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	PatientID         uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`
	ResponsibleUserID uuid.UUID `gorm:"type:uuid;not null" json:"responsible_user_id"`

	Type string `gorm:"not null" json:"type"`

	Symptoms            string `json:"symptoms,omitempty"`
	SymptomIntensity    *int   `gorm:"check:symptom_intensity >= 1 AND symptom_intensity <= 10" json:"symptom_intensity,omitempty"`
	SideEffects         string `json:"side_effects,omitempty"`
	SideEffectIntensity *int   `gorm:"check:side_effect_intensity >= 1 AND side_effect_intensity <= 10" json:"side_effect_intensity,omitempty"`

	TreatmentAdherence string `gorm:"check:treatment_adherence IN ('alta', 'media', 'baixa')" json:"treatment_adherence,omitempty"`
	Challenges         string `json:"challenges,omitempty"`
	Improvements       string `json:"improvements,omitempty"`
	AdditionalNotes    string `json:"additional_notes,omitempty"`

	Weight        *float64 `gorm:"type:decimal(5,2)" json:"weight,omitempty"`
	BloodPressure string   `json:"blood_pressure,omitempty"`
	HeartRate     *int     `json:"heart_rate,omitempty"`

	ExtraResponses map[string]interface{} `gorm:"type:jsonb" json:"extra_responses,omitempty"`

	Association     *Association `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Patient         *Patient     `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	ResponsibleUser *User        `gorm:"foreignKey:ResponsibleUserID" json:"responsible_user,omitempty"`
}

func (Anamnese) TableName() string {
	return "anamneses"
}

func (a *Anamnese) BeforeCreate(tx *gorm.DB) error {
	if a.Type == "" {
		a.Type = "inicial"
	}
	return nil
}