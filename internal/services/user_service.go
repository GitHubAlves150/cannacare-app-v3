// ================================================================
// PACOTE SERVICES - USER SERVICE (GESTÃO DE USUÁRIOS DA ASSOCIAÇÃO)
// ================================================================
// ⚠️ Este arquivo SUBSTITUI a versão anterior de user_service.go.
//
// Adiciona/gerencia usuários dentro de uma associação JÁ EXISTENTE.
// Diferente do AuthService.Register, que cria uma associação NOVA.
//
// REGRA DE NEGÓCIO - LIMITE POR PLANO:
//   O limite fica em association.UserLimit (coluna user_limit).
//   basic      -> 3
//   premium    -> 10
//   enterprise -> ilimitado (valor bem alto, ex: 999999)
//   Tudo definido por set_default_user_limit no banco.
//   A checagem aqui é feita ANTES de tentar inserir, pra devolver
//   uma mensagem amigável. A trigger enforce_user_limit_before_insert
//   no banco garante a regra mesmo se algo inserir fora da API.
// ================================================================

package services

import (
	"errors"
	"fmt"

	"cannacare-backend/internal/models"
	"cannacare-backend/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT USERSERVICE
// ================================================================
type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=200"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin secretaria acolhimento farmacia coordenacao"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin secretaria acolhimento farmacia coordenacao paciente"`
}

type AdminUserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// UserLimitInfo - usado pelo front-end para mostrar "3/3 usuários" etc.
type UserLimitInfo struct {
	Plan      string `json:"plan"`
	Limit     int    `json:"limit"`
	Used      int    `json:"used"`
	Available int    `json:"available"`
}

// ================================================================
// FUNÇÃO AUXILIAR: countActiveUsers + buscar associação
// ================================================================
func (s *UserService) usage(associationID uuid.UUID) (*models.Association, int64, error) {
	var association models.Association
	if err := s.db.Where("id = ?", associationID).First(&association).Error; err != nil {
		return nil, 0, err
	}

	var count int64
	if err := s.db.Model(&models.User{}).
		Where("association_id = ? AND is_active = true", associationID).
		Count(&count).Error; err != nil {
		return nil, 0, err
	}

	return &association, count, nil
}

// ================================================================
// FUNÇÃO GETUSAGE() - PÚBLICA
// ================================================================
func (s *UserService) GetUsage(associationID uuid.UUID) (*UserLimitInfo, error) {
	association, count, err := s.usage(associationID)
	if err != nil {
		return nil, err
	}
	available := association.UserLimit - int(count)
	if available < 0 {
		available = 0
	}
	return &UserLimitInfo{
		Plan:      association.Plan,
		Limit:     association.UserLimit,
		Used:      int(count),
		Available: available,
	}, nil
}

// ================================================================
// FUNÇÃO CREATE() - ADICIONAR USUÁRIO À ASSOCIAÇÃO
// ================================================================
func (s *UserService) Create(associationID uuid.UUID, req CreateUserRequest) (*AdminUserResponse, error) {
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("email inválido")
	}

	// --- Checar limite do plano ANTES de criar ---
	association, count, err := s.usage(associationID)
	if err != nil {
		return nil, err
	}
	if int(count) >= association.UserLimit {
		return nil, fmt.Errorf("limite de %d usuários do plano %s atingido. Faça upgrade para o plano premium para adicionar mais usuários", association.UserLimit, association.Plan)
	}

	// --- Email precisa ser único no sistema todo (login busca só por email) ---
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email já cadastrado")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("erro ao criptografar senha")
	}

	user := &models.User{
		AssociationID: associationID, // ← usa a associação do admin logado, NÃO cria uma nova
		Name:          req.Name,
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		Role:          req.Role,
		IsActive:      true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

// ================================================================
// FUNÇÃO LIST() - LISTAR USUÁRIOS DA ASSOCIAÇÃO
// ================================================================
func (s *UserService) List(associationID uuid.UUID) ([]AdminUserResponse, error) {
	var users []models.User
	if err := s.db.Where("association_id = ?", associationID).
		Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, err
	}

	responses := make([]AdminUserResponse, 0, len(users))
	for _, u := range users {
		responses = append(responses, *toUserResponse(&u))
	}
	return responses, nil
}

// ================================================================
// FUNÇÃO UPDATEROLE()
// ================================================================
func (s *UserService) UpdateRole(associationID, id uuid.UUID, newRole string) (*AdminUserResponse, error) {
	var user models.User
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}

	user.Role = newRole
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return toUserResponse(&user), nil
}

// ================================================================
// FUNÇÃO TOGGLESTATUS() - ATIVAR/DESATIVAR (sem precisar de body)
// ================================================================
// Desativar libera uma vaga no limite do plano básico sem apagar o
// histórico do usuário.
func (s *UserService) ToggleStatus(associationID, id uuid.UUID) (*AdminUserResponse, error) {
	var user models.User
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}

	// Se está REATIVANDO, precisa checar o limite de novo
	if !user.IsActive {
		association, count, err := s.usage(associationID)
		if err != nil {
			return nil, err
		}
		if int(count) >= association.UserLimit {
			return nil, fmt.Errorf("limite de %d usuários do plano %s atingido. Desative outro usuário ou faça upgrade", association.UserLimit, association.Plan)
		}
	}

	user.IsActive = !user.IsActive
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return toUserResponse(&user), nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================
func toUserResponse(user *models.User) *AdminUserResponse {
	return &AdminUserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}