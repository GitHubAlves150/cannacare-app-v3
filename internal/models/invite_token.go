// ================================================================
// MODEL INVITE_TOKEN (LINK DE PRIMEIRO ACESSO)
// ================================================================
// Usado para o admin definir a própria senha no primeiro acesso,
// em vez de o sistema mandar uma senha padrão por email.
//
// Guardamos só o HASH do token (SHA-256) — igual senha, nunca o
// valor puro fica salvo. O valor puro só existe no link do email.
// ================================================================

package models

import (
	"time"

	"github.com/google/uuid"
)

type InviteToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"type:varchar(64);unique;not null" json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (InviteToken) TableName() string {
	return "invite_tokens"
}