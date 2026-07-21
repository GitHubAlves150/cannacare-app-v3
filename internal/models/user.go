// ================================================================
// MODEL USER
// ================================================================
// Representa um usuário do sistema (administradores, staff, pacientes)
//
// Tabela: users
//
// RELACIONAMENTOS:
//   - Um usuário pode ser um paciente (relacionamento 1:1 com Patient)
//   - Um usuário pode criar/atualizar registros (auditoria)
//
// PERFIS (ROLES):
//   admin        - Acesso total ao sistema
//   coordenacao  - Gerenciar relatórios e aprovações
//   secretaria   - Cadastrar pacientes e documentos
//   acolhimento  - Realizar anamnese e acompanhamento
//   farmacia     - Gerenciar estoque e pedidos
//   paciente     - Acesso ao portal do paciente
//   cuidador     - Acesso ao portal em nome do paciente
// ================================================================

package models

import (
	"time"

	"gorm.io/gorm"
)

// User representa um usuário do sistema
// Esta tabela é a base para autenticação e controle de acesso
type User struct {
	// === CAMPOS BASE (EMBUTIDOS) ===
	BaseModel

	// === CAMPOS OBRIGATÓRIOS ===
	// Name é o nome completo do usuário
	Name string `gorm:"not null" json:"name"`

	// Email é único e usado para login
	Email string `gorm:"unique;not null" json:"email"`

	// PasswordHash armazena a senha criptografada (bcrypt)
	// O campo tem tag json:"-" para NUNCA ser exposto em respostas JSON
	PasswordHash string `gorm:"not null" json:"-"`

	// Role define as permissões do usuário no sistema
	// Valores: admin, coordenacao, secretaria, acolhimento, farmacia, paciente, cuidador
	Role string `gorm:"not null" json:"role"`

	// === CAMPOS OPCIONAIS ===
	// IsActive permite desativar um usuário sem deletá-lo
	IsActive bool `gorm:"default:true" json:"is_active"`

	// LastLoginAt registra a última vez que o usuário fez login
	// Usado para auditoria e estatísticas
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// === RELACIONAMENTOS ===
	// Patient - Relacionamento 1:1 com a tabela patients
	// Um usuário pode estar associado a um paciente (portal)
	Patient *Patient `gorm:"foreignKey:UserID" json:"patient,omitempty"`
}

// TableName define o nome da tabela no PostgreSQL
// Sempre usar plural no nome da tabela
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