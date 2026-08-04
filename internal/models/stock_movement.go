// ================================================================
// MODEL STOCK_MOVEMENT (MOVIMENTAÇÕES DE ESTOQUE)
// ================================================================
// ⚠️ Adicionado AssociationID para multi-tenancy.
// ================================================================

package models

import (
	"github.com/google/uuid"
)

type StockMovement struct {
	BaseModel

	// AssociationID - QUAL associação esta movimentação pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	ProductLotID uuid.UUID  `gorm:"type:uuid;not null" json:"product_lot_id"`
	OrderID      *uuid.UUID `gorm:"type:uuid" json:"order_id,omitempty"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`

	// Type - entrada, baixa_pedido, ajuste_manual, perda
	Type string `gorm:"not null" json:"type"`

	Quantity         int    `gorm:"not null" json:"quantity"`
	PreviousQuantity int    `gorm:"not null" json:"previous_quantity"`
	NewQuantity      int    `gorm:"not null" json:"new_quantity"`
	Notes            string `json:"notes,omitempty"`

	Association *Association `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	ProductLot  *ProductLot  `gorm:"foreignKey:ProductLotID" json:"product_lot,omitempty"`
	Order       *Order       `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	User        *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (StockMovement) TableName() string {
	return "stock_movements"
}