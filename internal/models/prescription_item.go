// ================================================================
// MODEL PRESCRIPTION_ITEM (ITENS DA RECEITA)
// ================================================================
// Detalha os medicamentos e dosagens recomendados em uma receita
//
// Tabela: prescription_items
//
// RELACIONAMENTOS:
//   - Pertence a uma receita (Prescription)
//   - Pertence a um produto (Product)
//
// EXEMPLO:
//   Receita: Dr. João para Maria
//   Item 1: Óleo CBD 10% - 5 gotas a cada 12 horas - 30ml
//   Item 2: Óleo CBD 20% - 3 gotas ao dormir - 30ml
// ================================================================

package models

import (
	"github.com/google/uuid"
)

// PrescriptionItem representa um item de uma receita médica
type PrescriptionItem struct {
	// === CAMPOS BASE ===
	BaseModel

	// === CHAVES ESTRANGEIRAS ===
	// PrescriptionID - Receita a qual este item pertence
	PrescriptionID uuid.UUID `gorm:"type:uuid;not null" json:"prescription_id"`

	// ProductID - Produto/medicamento prescrito
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`

	// === DETALHES DA PRESCRIÇÃO ===
	// DosageInstructions - Como o paciente deve tomar
	// Ex: "5 gotas a cada 12 horas, sublingual"
	DosageInstructions string `gorm:"not null" json:"dosage_instructions"`

	// QuantityRecommended - Quantidade recomendada do produto
	// Ex: 1 (um frasco), 2 (dois frascos), etc
	QuantityRecommended int `gorm:"not null" json:"quantity_recommended"`

	// === RELACIONAMENTOS ===
	Prescription *Prescription `gorm:"foreignKey:PrescriptionID" json:"prescription,omitempty"`
	Product     *Product      `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName define o nome da tabela
func (PrescriptionItem) TableName() string {
	return "prescription_items"
}