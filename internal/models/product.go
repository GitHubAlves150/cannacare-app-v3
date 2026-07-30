// ================================================================
// CANNACARE - MODEL PRODUCT (PRODUTO)
// ================================================================
// Catálogo de produtos/medicamentos disponíveis para venda
//
// TABELA: products
//
// RELACIONAMENTOS:
//   - Cada produto pertence a UMA associação (association_id)
//   - Tem muitos lotes (ProductLot)
//   - Aparece em muitos itens de receita (PrescriptionItem)
//   - Aparece em muitos itens de pedido (OrderItem)
//
// REGRAS DE NEGÓCIO:
//   - Produtos podem ser ativos/inativos (is_active)
//   - Cada produto tem um preço unitário (unit_price)
//   - MinStockAlert define quando gerar alerta de estoque baixo
// ================================================================

package models

import (
	"github.com/google/uuid"
)

// Product representa um produto/medicamento no catálogo
type Product struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVE ESTRANGEIRA PARA MULTI-TENANCY ===
	// AssociationID - QUAL associação este produto pertence
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	// === DADOS DO PRODUTO ===
	// Name - Nome comercial do produto
	Name string `gorm:"not null" json:"name"`

	// Description - Descrição detalhada do produto
	Description string `json:"description,omitempty"`

	// UnitPrice - Preço unitário do produto
	UnitPrice float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`

	// MinStockAlert - Nível mínimo de estoque para gerar alerta
	MinStockAlert int `gorm:"default:10" json:"min_stock_alert"`

	// IsActive - Se o produto está ativo no catálogo
	IsActive bool `gorm:"default:true" json:"is_active"`

	// === RELACIONAMENTOS ===
	Association      *Association       `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
	Lots             []ProductLot       `gorm:"foreignKey:ProductID" json:"lots,omitempty"`
	PrescriptionItems []PrescriptionItem `gorm:"foreignKey:ProductID" json:"prescription_items,omitempty"`
}

// TableName define o nome da tabela
func (Product) TableName() string {
	return "products"
}