// ================================================================
// MODEL PRODUCT_LOT (LOTES DE PRODUTOS)
// ================================================================
// Controle de estoque por lote (produto + lote + validade)
//
// Tabela: product_lots
//
// RELACIONAMENTOS:
//   - Pertence a um produto (Product)
//   - Tem muitas movimentações de estoque (StockMovement)
//   - Aparece em muitos itens de pedido (OrderItem)
//
// POR QUE CONTROLAR POR LOTE?
// 1. Rastreabilidade - Saber de onde veio cada produto
// 2. Validade - Controlar produtos que estão vencendo
// 3. Recall - Em caso de problema, saber quais lotes foram afetados
// 4. Custo - Controlar custo por lote (quando comprou, quanto pagou)
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductLot representa um lote de um produto
type ProductLot struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// ProductID - Produto ao qual este lote pertence
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`

	// ReceivedBy - Usuário que recebeu o lote
	ReceivedBy *uuid.UUID `gorm:"type:uuid" json:"received_by,omitempty"`

	// === DADOS DO LOTE ===
	// LotNumber - Número do lote (identificação única)
	// Ex: "LOTE-001", "LOTE-2024-001"
	LotNumber string `gorm:"not null" json:"lot_number"`

	// ExpirationDate - Data de validade do lote
	// Produtos vencidos não podem ser vendidos
	ExpirationDate time.Time `gorm:"not null" json:"expiration_date"`

	// CurrentQuantity - Quantidade atual em estoque
	// Atualizada automaticamente com movimentações (entrada/saída)
	CurrentQuantity int `gorm:"not null;default:0" json:"current_quantity"`

	// InitialQuantity - Quantidade inicial do lote
	// Útil para saber quanto entrou no total
	InitialQuantity int `gorm:"not null;default:0" json:"initial_quantity"`

	// Supplier - Fornecedor do lote
	Supplier string `json:"supplier,omitempty"`

	// PurchaseDate - Data da compra
	PurchaseDate *time.Time `json:"purchase_date,omitempty"`

	// PurchasePrice - Preço de compra (custo)
	PurchasePrice float64 `gorm:"type:decimal(10,2)" json:"purchase_price,omitempty"`

	// ReceivedAt - Data/hora em que o lote foi recebido
	ReceivedAt *time.Time `json:"received_at,omitempty"`

	// === RELACIONAMENTOS ===
	Product         *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	StockMovements  []StockMovement   `gorm:"foreignKey:ProductLotID" json:"stock_movements,omitempty"`
	OrderItems      []OrderItem       `gorm:"foreignKey:ProductLotID" json:"order_items,omitempty"`
}

// TableName define o nome da tabela
func (ProductLot) TableName() string {
	return "product_lots"
}