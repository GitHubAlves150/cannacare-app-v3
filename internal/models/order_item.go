// ================================================================
// MODEL ORDER_ITEM (ITENS DO PEDIDO)
// ================================================================
// Produtos específicos que compõem um pedido
//
// Tabela: order_items
//
// RELACIONAMENTOS:
//   - Pertence a um pedido (Order)
//   - Pertence a um lote específico (ProductLot)
//
// CÁLCULO AUTOMÁTICO:
//   total_price = quantity * unit_price (campo GENERATED ALWAYS)
// ================================================================

package models

import (
	"github.com/google/uuid"
)

// OrderItem representa um item de um pedido
type OrderItem struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// OrderID - Pedido ao qual este item pertence
	OrderID uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`

	// ProductLotID - Lote específico do produto
	ProductLotID uuid.UUID `gorm:"type:uuid;not null" json:"product_lot_id"`

	// === DADOS DO ITEM ===
	// Quantity - Quantidade solicitada
	Quantity int `gorm:"not null" json:"quantity"`

	// UnitPrice - Preço unitário no momento do pedido
	UnitPrice float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`

	// TotalPrice - Preço total (quantity * unit_price)
	// CALCULADO AUTOMATICAMENTE PELO BANCO (GENERATED ALWAYS)
	TotalPrice float64 `gorm:"type:decimal(10,2);generated:always as (quantity * unit_price) stored" json:"total_price"`

	// === RELACIONAMENTOS ===
	Order      *Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	ProductLot *ProductLot `gorm:"foreignKey:ProductLotID" json:"product_lot,omitempty"`
}

// TableName define o nome da tabela
func (OrderItem) TableName() string {
	return "order_items"
}