// ================================================================
// PACOTE HANDLERS - AUTH HANDLER
// ================================================================
// Camada HTTP que lida com as requisições de autenticação.
// ================================================================

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"cannacare-backend/internal/services"  // ← CORRIGIR: use o nome do seu módulo

	"github.com/go-playground/validator/v10"
)

// ================================================================
// STRUCT AUTHHANDLER
// ================================================================
type AuthHandler struct {
	authService *services.AuthService
	validator   *validator.Validate
}

// ================================================================
// FUNÇÃO NEWAUTHHANDLER()
// ================================================================
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// ================================================================
// FUNÇÕES AUXILIARES DE RESPOSTA
// ================================================================

// sendSuccess envia uma resposta de sucesso em JSON
func sendSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// sendError envia uma resposta de erro em JSON
func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// ================================================================
// HANDLER: REGISTER
// ================================================================
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req services.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		sendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		if strings.Contains(err.Error(), "email já cadastrado") {
			sendError(w, http.StatusConflict, err.Error())
			return
		}
		log.Printf("❌ Erro ao registrar usuário: %v", err)
		sendError(w, http.StatusInternalServerError, "erro interno ao registrar usuário")
		return
	}

	sendSuccess(w, http.StatusCreated, user)
}

// ================================================================
// HANDLER: LOGIN
// ================================================================
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req services.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		sendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	authResponse, err := h.authService.Login(req)
	if err != nil {
		if strings.Contains(err.Error(), "email ou senha incorretos") {
			sendError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if strings.Contains(err.Error(), "usuário desativado") {
			sendError(w, http.StatusForbidden, err.Error())
			return
		}
		log.Printf("❌ Erro ao fazer login: %v", err)
		sendError(w, http.StatusInternalServerError, "erro interno ao fazer login")
		return
	}

	sendSuccess(w, http.StatusOK, authResponse)
}