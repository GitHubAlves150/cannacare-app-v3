// ================================================================
// PACOTE HANDLERS - ADMIN USER HANDLER
// ================================================================
// ⚠️ NOVO ARQUIVO. Implementa exatamente os endpoints que o
// front-end (lib/api/users.ts) já chama:
//
//   GET   /api/admin/users             -> List
//   POST  /api/admin/users             -> Create   (novo, ainda não tinha)
//   PATCH /api/admin/users/{id}/role   -> UpdateRole
//   PATCH /api/admin/users/{id}/status -> ToggleStatus (sem body)
// ================================================================

package handlers

import (
	"encoding/json"
	"net/http"

	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type AdminUserHandler struct {
	userService *services.UserService
	validator   *validator.Validate
}

func NewAdminUserHandler(userService *services.UserService) *AdminUserHandler {
	return &AdminUserHandler{
		userService: userService,
		validator:   validator.New(),
	}
}

func (h *AdminUserHandler) extractAssociationID(r *http.Request) (uuid.UUID, error) {
	associationIDStr, _ := r.Context().Value(middleware.AssociationIDKey).(string)
	if associationIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(associationIDStr)
}

// ================================================================
// GET /api/admin/users
// ================================================================
// Front-end espera: response.data.items
func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	users, err := h.userService.List(associationID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar usuários")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": users,
	})
}

// ================================================================
// POST /api/admin/users
// ================================================================
// Adiciona um usuário à MESMA associação do admin logado.
// Body: {"name":"...","email":"...","password":"...","role":"secretaria"}
func (h *AdminUserHandler) Create(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	var req services.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	user, err := h.userService.Create(associationID, req)
	if err != nil {
		// erro de negócio (limite atingido, email duplicado) -> 400
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, user)
}

// ================================================================
// PATCH /api/admin/users/{id}/role
// ================================================================
// Body: {"role":"secretaria"}
func (h *AdminUserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req services.UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	user, err := h.userService.UpdateRole(associationID, id, req.Role)
	if err != nil {
		if err.Error() == "usuário não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, user)
}

// ================================================================
// PATCH /api/admin/users/{id}/status
// ================================================================
// O front-end chama SEM body - só inverte o status atual.
func (h *AdminUserHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	user, err := h.userService.ToggleStatus(associationID, id)
	if err != nil {
		if err.Error() == "usuário não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, user)
}