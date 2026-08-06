// ================================================================
// PACOTE HANDLERS - PROFILE HANDLER
// ================================================================
// Endpoints do próprio usuário logado gerenciar os dados dele.
// Diferente do billing_handler.go (dados da associação) e do
// admin_user_handler.go (admin gerenciando OUTROS usuários) — aqui
// é sempre "eu mesmo".
//
//   GET  /api/users/me           - ver meus dados
//   PUT  /api/users/me           - editar nome/email
//   PUT  /api/users/me/password  - trocar senha
// ================================================================

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/models"
	"cannacare-backend/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ProfileHandler struct {
	db *gorm.DB
}

func NewProfileHandler(db *gorm.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

func (h *ProfileHandler) extractUserID(r *http.Request) (uuid.UUID, error) {
	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	if userIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(userIDStr)
}

type profileResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

func toProfileResponse(u *models.User) profileResponse {
	return profileResponse{
		ID:        u.ID.String(),
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ================================================================
// GET /api/users/me
// ================================================================
func (h *ProfileHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil || userID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "usuário não identificado")
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("❌ erro ao buscar perfil: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao carregar perfil")
		return
	}

	utils.SendSuccess(w, http.StatusOK, toProfileResponse(&user))
}

type updateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=3,max=200"`
	Email string `json:"email" validate:"omitempty,email"`
}

// ================================================================
// PUT /api/users/me
// ================================================================
func (h *ProfileHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil || userID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "usuário não identificado")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "usuário não encontrado")
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if req.Email != "" && req.Email != user.Email {
		if !utils.IsValidEmail(req.Email) {
			utils.SendError(w, http.StatusBadRequest, "email inválido")
			return
		}
		var existing models.User
		if err := h.db.Where("email = ? AND id != ?", req.Email, userID).First(&existing).Error; err == nil {
			utils.SendError(w, http.StatusBadRequest, "este email já está em uso")
			return
		} else if err != gorm.ErrRecordNotFound {
			utils.SendError(w, http.StatusInternalServerError, "erro ao validar email")
			return
		}
		user.Email = req.Email
	}

	if err := h.db.Save(&user).Error; err != nil {
		log.Printf("❌ erro ao atualizar perfil: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao atualizar perfil")
		return
	}

	utils.SendSuccess(w, http.StatusOK, toProfileResponse(&user))
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// ================================================================
// PUT /api/users/me/password
// ================================================================
func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil || userID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "usuário não identificado")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if len(req.NewPassword) < 6 {
		utils.SendError(w, http.StatusBadRequest, "a nova senha precisa ter pelo menos 6 caracteres")
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "usuário não encontrado")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		utils.SendError(w, http.StatusBadRequest, "senha atual incorreta")
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao processar nova senha")
		return
	}

	user.PasswordHash = string(newHash)
	if err := h.db.Save(&user).Error; err != nil {
		log.Printf("❌ erro ao trocar senha: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao trocar senha")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{"message": "senha alterada com sucesso"})
}