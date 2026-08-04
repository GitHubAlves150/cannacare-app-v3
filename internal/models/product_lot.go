// ================================================================
// MODEL PRODUCT_LOT (LOTES DE PRODUTOS)
// ================================================================
// ⚠️ Adicionado AssociationID para multi-tenancy.
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductLot struct {
	BaseModel

	// AssociationID - QUAL associação este lote pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	ProductID  uuid.UUID  `gorm:"type:uuid;not null" json:"product_id"`
	ReceivedBy *uuid.UUID `gorm:"type:uuid" json:"received_by,omitempty"`

	LotNumber       string    `gorm:"not null" json:"lot_number"`
	ExpirationDate  time.Time `gorm:"not null" json:"expiration_date"`
	CurrentQuantity int       `gorm:"not null;default:0" json:"current_quantity"`
	InitialQuantity int       `gorm:"not null;default:0" json:"initial_quantity"`
	Supplier        string    `json:"supplier,omitempty"`
	PurchaseDate    *time.Time `json:"purchase_date,omitempty"`
	PurchasePrice   float64    `gorm:"type:decimal(10,2)" json:"purchase_price,omitempty"`
	ReceivedAt      *time.Time `json:"received_at,omitempty"`

	Association    *Association    `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Product        *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	StockMovements []StockMovement `gorm:"foreignKey:ProductLotID" json:"stock_movements,omitempty"`
	OrderItems     []OrderItem     `gorm:"foreignKey:ProductLotID" json:"order_items,omitempty"`
}

func (ProductLot) TableName() string {
	return "product_lots"
}