// ================================================================
// MODEL ORDER_ITEM (ITENS DO PEDIDO)
// ================================================================
// ⚠️ Adicionado AssociationID para multi-tenancy.
// ================================================================

package models

import (
	"github.com/google/uuid"
)

type OrderItem struct {
	BaseModel

	// AssociationID - QUAL associação este item pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	OrderID      uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`
	ProductLotID uuid.UUID `gorm:"type:uuid;not null" json:"product_lot_id"`

	Quantity  int     `gorm:"not null" json:"quantity"`
	UnitPrice float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`

	// TotalPrice - CALCULADO AUTOMATICAMENTE PELO BANCO (GENERATED ALWAYS)
	// "->" = GORM só LÊ esse campo, nunca tenta escrever (senão o Postgres
	// rejeita, já que é uma coluna gerada pelo próprio banco)
	TotalPrice float64 `gorm:"->;type:decimal(10,2)" json:"total_price"`

	Association *Association `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Order       *Order       `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	ProductLot  *ProductLot  `gorm:"foreignKey:ProductLotID" json:"product_lot,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_items"
}