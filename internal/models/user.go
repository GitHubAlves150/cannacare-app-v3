// ================================================================
// CANNACARE - MODEL USER (USUÁRIO)
// ================================================================
// Representa um usuário do sistema (staff da associação)
//
// TABELA: users
//
// RELACIONAMENTOS:
//   - Cada usuário pertence a UMA associação (association_id)
//   - Um usuário pode ser um paciente (relacionamento 1:1 com Patient)
//
// PERFIS (ROLES):
//   admin        - Acesso total ao sistema
//   coordenacao  - Gerenciar relatórios e aprovações
//   secretaria   - Cadastrar pacientes e documentos
//   acolhimento  - Realizar anamnese e acompanhamento
//   farmacia     - Gerenciar estoque e pedidos
//
// FLUXO DE AUTENTICAÇÃO:
//   1. Usuário envia email + senha
//   2. Backend valida credenciais
//   3. Backend gera JWT com association_id
//   4. Todas as requisições futuras incluem association_id no token
//   5. Backend filtra todas as queries por association_id
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User representa um usuário do sistema
type User struct {
	// === CAMPOS BASE (EMBUTIDOS) ===
	BaseModel

	// === CHAVE ESTRANGEIRA PARA MULTI-TENANCY ===
	// AssociationID - QUAL associação este usuário pertence
	// ⚠️ NUNCA deixar nulo! Todo usuário pertence a uma associação
	// IMPORTANTE: Esta coluna É O CORAÇÃO DO MULTI-TENANCY
	//
	// COMO FUNCIONA:
	//   1. Quando o usuário faz login, o association_id é extraído do banco
	//   2. O association_id é incluído no JWT (token)
	//   3. O middleware extrai o association_id de CADA requisição
	//   4. Todas as queries SQL filtram por association_id
	//
	// EXEMPLO DE QUERY SEGURA:
	//   SELECT * FROM patients WHERE association_id = ?
	//
	// EXEMPLO DE QUERY PERIGOSA (NUNCA FAZER):
	//   SELECT * FROM patients  -- Isso MOSTRARIA TODOS OS PACIENTES DE TODAS ASSOCIAÇÕES!
	AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`

	// === CAMPOS OBRIGATÓRIOS ===
	// Name é o nome completo do usuário
	Name string `gorm:"not null" json:"name"`

	// Email é único por associação (um mesmo email pode existir em associações diferentes)
	// A constraint UNIQUE(email, association_id) garante isso
	Email string `gorm:"not null" json:"email"`

	// PasswordHash armazena a senha criptografada com bcrypt
	// O campo tem tag json:"-" para NUNCA ser exposto em respostas JSON
	PasswordHash string `gorm:"not null" json:"-"`

	// Role define as permissões do usuário no sistema
	// admin, coordenacao, secretaria, acolhimento, farmacia
	Role string `gorm:"not null" json:"role"`

	// === CAMPOS OPCIONAIS ===
	// IsActive permite desativar um usuário sem deletá-lo
	IsActive bool `gorm:"default:true" json:"is_active"`

	// LastLoginAt registra a última vez que o usuário fez login
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// === RELACIONAMENTOS ===
	// Patient - Relacionamento 1:1 com a tabela patients
	Patient *Patient `gorm:"foreignKey:UserID" json:"patient,omitempty"`

	// Association - A associação à qual este usuário pertence
	Association *Association `gorm:"foreignKey:AssociationID" json:"association,omitempty"`
}

// TableName define o nome da tabela no PostgreSQL
func (User) TableName() string {
	return "users"
}

// BeforeCreate é um hook do GORM executado ANTES de criar um registro
// Usado para validações ou preenchimento automático de campos
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Se não houver role definida, usa 'paciente' como padrão
	if u.Role == "" {
		u.Role = "paciente"
	}
	return nil
}