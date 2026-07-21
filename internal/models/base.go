// ================================================================
// PACOTE MODELS
// ================================================================
// Define a estrutura base para todos os models do sistema.
//
// Esta struct é EMBUTIDA em todos os outros models para
// padronizar campos comuns como ID, CreatedAt, UpdatedAt e DeletedAt.
//
// POR QUE USAR UM BASE MODEL?
// 1. DRY (Don't Repeat Yourself) - Evita repetição de campos
// 2. Padronização - Todos os models têm os mesmos campos base
// 3. Facilidade de manutenção - Mudar um campo apenas em um lugar
// 4. GORM reconhece automaticamente os campos padrão
//
// EXEMPLO:
//   type Patient struct {
//       BaseModel           // Embute os campos base
//       Name string         // Campos específicos do paciente
//       CPF  string
//   }
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
// BaseModel contém os campos comuns a todas as tabelas do sistema.
//
// CAMPOS:
//
//	ID        - UUID único (chave primária)
//	CreatedAt - Data/hora de criação (automático)
//	UpdatedAt - Data/hora da última atualização (automático)
//	DeletedAt - Data/hora da exclusão lógica (soft delete)
type BaseModel struct {
	// ID é a chave primária de todas as tabelas
	// Usa UUID (Universally Unique Identifier) em vez de inteiro sequencial
	// Vantagens: Não é sequencial, não expõe quantidade de registros,
	//            É único globalmente, seguro para APIs públicas
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
