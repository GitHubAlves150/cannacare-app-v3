// ================================================================
// MODEL ORDER (PEDIDOS)
// ================================================================
// ⚠️ Adicionado AssociationID para multi-tenancy.
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	BaseModel

	// AssociationID - QUAL associação este pedido pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	PatientID      uuid.UUID `gorm:"type:uuid;not null" json:"patient_id"`
	PrescriptionID uuid.UUID `gorm:"type:uuid;not null" json:"prescription_id"`

	Status          string    `gorm:"not null;default:'pendente'" json:"status"`
	StatusUpdatedAt time.Time `gorm:"autoUpdateTime" json:"status_updated_at"`

	TotalAmount float64   `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Notes       string    `json:"notes,omitempty"`
	OrderDate   time.Time `gorm:"autoCreateTime" json:"order_date"`

	ShippingCarrier  string  `json:"shipping_carrier,omitempty"`
	TrackingCode     string  `json:"tracking_code,omitempty"`
	ShippingLabelURL string  `json:"shipping_label_url,omitempty"`
	ShippingCost     float64 `gorm:"type:decimal(10,2);default:0" json:"shipping_cost"`

	Association  *Association  `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Patient      *Patient      `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Prescription *Prescription `gorm:"foreignKey:PrescriptionID" json:"prescription,omitempty"`
	Items        []OrderItem   `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Payment      *Payment      `gorm:"foreignKey:OrderID" json:"payment,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}