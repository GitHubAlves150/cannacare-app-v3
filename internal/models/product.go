// ================================================================
// MODEL PRODUCT (PRODUTOS / MEDICAMENTOS)
// ================================================================
// Catálogo de produtos/medicamentos disponíveis para venda
//
// Tabela: products
//
// RELACIONAMENTOS:
//   - Tem muitos lotes (ProductLot)
//   - Aparece em muitos itens de receita (PrescriptionItem)
//   - Aparece em muitos itens de pedido (OrderItem)
//
// EXEMPLOS DE PRODUTOS:
//   - Óleo CBD Full Spectrum 10% - 30ml
//   - Óleo CBD Full Spectrum 20% - 30ml
//   - Óleo CBD Isolado 5% - 30ml
//   - Cápsulas CBD 25mg - 60 unidades
// ================================================================

package models

// Product representa um produto/medicamento no catálogo
type Product struct {
	// === CAMPOS BASE ===
	BaseModel

	// === DADOS DO PRODUTO ===
	// Name - Nome comercial do produto
	// Ex: "Óleo CBD Full Spectrum 10% - 30ml"
	Name string `gorm:"not null" json:"name"`

	// Description - Descrição detalhada do produto
	// Ex: "Óleo com 10% de CBD, 30ml, uso sublingual"
	Description string `json:"description,omitempty"`

	// UnitPrice - Preço unitário do produto
	// Usar DECIMAL para valores monetários (evita problemas com floats)
	UnitPrice float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`

	// MinStockAlert - Nível mínimo de estoque para gerar alerta
	// Quando a quantidade em estoque for menor ou igual a este valor,
	// um alerta é disparado
	MinStockAlert int `gorm:"default:10" json:"min_stock_alert"`

	// IsActive - Se o produto está ativo no catálogo
	IsActive bool `gorm:"default:true" json:"is_active"`

	// === RELACIONAMENTOS ===
	Lots            []ProductLot      `gorm:"foreignKey:ProductID" json:"lots,omitempty"`
	PrescriptionItems []PrescriptionItem `gorm:"foreignKey:ProductID" json:"prescription_items,omitempty"`
}

// TableName define o nome da tabela
func (Product) TableName() string {
	return "products"
}