// ================================================================
// MODEL NOTIFICATION (NOTIFICAÇÕES)
// ================================================================
// ⚠️ Adicionado AssociationID para multi-tenancy.
// Ainda não há código criando notificações de verdade — isso deixa
// o model pronto para quando essa feature for implementada.
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	BaseModel

	// AssociationID - QUAL associação esta notificação pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	PatientID *uuid.UUID `gorm:"type:uuid" json:"patient_id,omitempty"`

	Type      string     `gorm:"not null" json:"type"`
	Title     string     `gorm:"not null" json:"title"`
	Message   string     `gorm:"not null" json:"message"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	ActionURL string     `json:"action_url,omitempty"`

	Association *Association `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	User        *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Patient     *Patient     `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}