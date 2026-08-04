// ================================================================
// CANNACARE - BASE MODEL
// ================================================================
// Este arquivo define a estrutura BASE que será EMBUTIDA em todos
// os outros models do sistema.
//
// POR QUE USAR UM BASE MODEL?
// 1. DRY (Don't Repeat Yourself) - Não repetimos campos em todo lugar
// 2. Padronização - Todos os models têm ID, timestamps e soft delete
// 3. Facilidade de manutenção - Mudamos em um só lugar
// 4. O GORM reconhece automaticamente os campos CreatedAt, UpdatedAt, DeletedAt
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT BASEMODEL
// ================================================================
// Todos os models do sistema vão EMBUTIR esta struct.
// Isso significa que todos terão os campos:
//   - ID: UUID único (chave primária)
//   - CreatedAt: Data de criação (automático)
//   - UpdatedAt: Data da última atualização (automático)
//   - DeletedAt: Data de exclusão lógica (soft delete)
//
// EXEMPLO DE USO:
//   type Patient struct {
//       BaseModel  // ← EMBUTE os campos base
//       Name string // ← Campos específicos
//   }
// ================================================================

type BaseModel struct {
	// ID é a chave primária de todas as tabelas
	// Usamos UUID (Universally Unique Identifier) em vez de inteiro sequencial
	// VANTAGENS DO UUID:
	//   - Não é sequencial (não expõe quantos registros existem)
	//   - É único globalmente (pode ser gerado em qualquer lugar)
	//   - Seguro para APIs públicas (não dá para adivinhar o próximo ID)
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	// CreatedAt é preenchido automaticamente pelo GORM
	// Quando um registro é criado, o GORM insere a data/hora atual
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// UpdatedAt é atualizado automaticamente pelo GORM
	// Sempre que um registro é modificado, o GORM atualiza este campo
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// DeletedAt é usado para "soft delete" (exclusão lógica)
	// Quando um registro é "deletado", apenas este campo é preenchido
	// O registro permanece no banco, mas é ocultado nas consultas normais
	// O GORM automaticamente filtra registros com DeletedAt != NULL
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}