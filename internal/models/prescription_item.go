// ================================================================
// MODEL PRESCRIPTION_ITEM (ITENS DA RECEITA)
// ================================================================
// ⚠️ Adicionado AssociationID para multi-tenancy.
// ================================================================

package models

import (
	"github.com/google/uuid"
)

type PrescriptionItem struct {
	BaseModel

	// AssociationID - QUAL associação este item pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	PrescriptionID uuid.UUID `gorm:"type:uuid;not null" json:"prescription_id"`
	ProductID      uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`

	DosageInstructions  string `gorm:"not null" json:"dosage_instructions"`
	QuantityRecommended int    `gorm:"not null" json:"quantity_recommended"`

	Association  *Association  `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Prescription *Prescription `gorm:"foreignKey:PrescriptionID" json:"prescription,omitempty"`
	Product      *Product      `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (PrescriptionItem) TableName() string {
	return "prescription_items"
}