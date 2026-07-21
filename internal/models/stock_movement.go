// ================================================================
// MODEL STOCK_MOVEMENT (MOVIMENTAÇÕES DE ESTOQUE)
// ================================================================
// Histórico detalhado de todas as movimentações de estoque
//
// Tabela: stock_movements
//
// TIPOS DE MOVIMENTAÇÃO (type):
//   entrada     - Entrada de produtos (compra)
//   baixa_pedido - Saída por pedido (venda)
//   ajuste_manual - Ajuste manual de estoque (perda, correção)
//   perda       - Perda de produtos (vencimento, dano)
//   devolucao   - Devolução de produtos
//
// POR QUE REGISTRAR MOVIMENTAÇÕES?
// 1. Auditoria - Saber exatamente o que aconteceu com cada produto
// 2. Rastreabilidade - Rastrear de onde veio e para onde foi
// 3. Relatórios - Gerar relatórios de movimentação
// 4. Controle - Evitar fraudes e erros
// ================================================================

package models

import (
	"github.com/google/uuid"
)

// StockMovement representa uma movimentação de estoque
type StockMovement struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// ProductLotID - Lote afetado pela movimentação
	ProductLotID uuid.UUID `gorm:"type:uuid;not null" json:"product_lot_id"`

	// OrderID - Pedido relacionado (se for baixa de pedido)
	OrderID *uuid.UUID `gorm:"type:uuid" json:"order_id,omitempty"`

	// UserID - Usuário que realizou a movimentação
	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`

	// === DADOS DA MOVIMENTAÇÃO ===
	// Type - Tipo de movimentação
	// entrada, baixa_pedido, ajuste_manual, perda, devolucao
	Type string `gorm:"not null" json:"type"`

	// Quantity - Quantidade movimentada
	// Positiva para entrada, negativa para saída
	Quantity int `gorm:"not null" json:"quantity"`

	// PreviousQuantity - Quantidade ANTES da movimentação
	// Útil para auditoria e histórico
	PreviousQuantity int `gorm:"not null" json:"previous_quantity"`

	// NewQuantity - Quantidade DEPOIS da movimentação
	// Deve ser igual a PreviousQuantity + Quantity
	NewQuantity int `gorm:"not null" json:"new_quantity"`

	// Notes - Observações adicionais
	Notes string `json:"notes,omitempty"`

	// === RELACIONAMENTOS ===
	ProductLot *ProductLot `gorm:"foreignKey:ProductLotID" json:"product_lot,omitempty"`
	Order      *Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName define o nome da tabela
func (StockMovement) TableName() string {
	return "stock_movements"
}