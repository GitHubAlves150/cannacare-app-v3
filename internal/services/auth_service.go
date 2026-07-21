// ================================================================
// PACOTE SERVICES - AUTH SERVICE
// ================================================================
// Camada de serviço responsável pela lógica de autenticação.
// ================================================================

package services

import (

	"cannacare-backend/internal/models"
	"cannacare-backend/pkg/jwt"
	"cannacare-backend/internal/utils"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT AUTHSERVICE
// ================================================================
type AuthService struct {
	db         *gorm.DB
	jwtService *jwt.JWTService
}

// ================================================================
// FUNÇÃO NEWAUTHSERVICE()
// ================================================================
func NewAuthService(db *gorm.DB, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		db:         db,
		jwtService: jwtService,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=50"`
	Role     string `json:"role" validate:"omitempty,oneof=admin secretaria acolhimento farmacia coordenacao paciente cuidador"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token     string        `json:"token"`
	TokenType string        `json:"token_type"`
	ExpiresIn int64         `json:"expires_in"`
	User      *UserResponse `json:"user"`
}

type UserResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ================================================================
// FUNÇÃO REGISTER()
// ================================================================
func (s *AuthService) Register(req RegisterRequest) (*UserResponse, error) {
	// --- 1. Validar email ---
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("email inválido")
	}

	// --- 2. Verificar se o email já existe ---
	var existingUser models.User
	result := s.db.Where("email = ?", req.Email).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("email já cadastrado")
	} else if result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	// --- 3. Definir role padrão ---
	role := req.Role
	if role == "" {
		role = "paciente"
	}

	// --- 4. Criptografar a senha ---
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("erro ao criptografar senha")
	}

	// --- 5. Criar o usuário ---
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		IsActive:     true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}, nil
}

// ================================================================
// FUNÇÃO LOGIN()
// ================================================================
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	// --- 1. Validar email ---
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("email inválido")
	}

	// --- 2. Buscar usuário pelo email ---
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("email ou senha incorretos")
		}
		return nil, err
	}

	// --- 3. Verificar se o usuário está ativo ---
	if !user.IsActive {
		return nil, errors.New("usuário desativado, contate o administrador")
	}

	// --- 4. Comparar a senha com o hash ---
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("email ou senha incorretos")
	}

	// --- 5. Gerar token JWT ---
	token, err := s.jwtService.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, errors.New("erro ao gerar token de autenticação")
	}

	// --- 6. Atualizar LastLoginAt ---
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	// --- 7. Calcular tempo de expiração ---
	expiresInSeconds := int64(s.jwtService.ExpiresIn().Seconds())

	return &AuthResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: expiresInSeconds,
		User: &UserResponse{
			ID:          user.ID.String(),
			Name:        user.Name,
			Email:       user.Email,
			Role:        user.Role,
			IsActive:    user.IsActive,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}

// ❌ REMOVER: func isValidEmail() (já está no utils)
// ❌ REMOVER: func validateDoctorData() (está no doctor_service.go)
