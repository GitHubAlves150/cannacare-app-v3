// ================================================================
// MODEL PATIENT_STATUS_HISTORY (HISTÓRICO DE STATUS)
// ================================================================
// ⚠️ Adicionado AssociationID para multi-tenancy.
// ================================================================

package models

import (
	"github.com/google/uuid"
)

type PatientStatusHistory struct {
	BaseModel

	// AssociationID - QUAL associação este histórico pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	PatientID uuid.UUID  `gorm:"type:uuid;not null" json:"patient_id"`
	ChangedBy *uuid.UUID `gorm:"type:uuid" json:"changed_by,omitempty"`

	OldStatus string `json:"old_status,omitempty"`
	NewStatus string `gorm:"not null" json:"new_status"`
	Reason    string `json:"reason,omitempty"`

	Association   *Association `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Patient       *Patient     `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	ChangedByUser *User        `gorm:"foreignKey:ChangedBy" json:"changed_by_user,omitempty"`
}

func (PatientStatusHistory) TableName() string {
	return "patient_status_history"
}