// ================================================================
// PACOTE MIDDLEWARE - AUTH
// ================================================================
// Middleware para proteger rotas com autenticação JWT.
//
// COMO FUNCIONA:
// 1. O cliente envia o token no header Authorization
// 2. O middleware extrai e valida o token
// 3. Se for válido, adiciona os dados do usuário no contexto
// 4. Se for inválido, retorna 401 Unauthorized
//
// HEADER ESPERADO:
//   Authorization: Bearer <token>
//
// CONTEXTO ADICIONADO:
//   - user_id: ID do usuário
//   - user_email: Email do usuário
//   - user_role: Role do usuário
// ================================================================

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"cannacare-backend/pkg/jwt"
)

// ================================================================
// CONTEXT KEYS
// ================================================================
// Chaves para armazenar dados no contexto da requisição
type ContextKey string

const (
	UserIDKey    ContextKey = "user_id"
	UserEmailKey ContextKey = "user_email"
	UserRoleKey  ContextKey = "user_role"
)

// ================================================================
// MIDDLEWARE: AUTHMIDDLEWARE
// ================================================================
// Verifica se a requisição possui um token JWT válido.
//
// FLUXO:
// 1. Extrai o token do header Authorization
// 2. Valida o token com o JWTService
// 3. Se válido, adiciona os dados no contexto
// 4. Chama o próximo handler
// 5. Se inválido, retorna 401 Unauthorized
//
// COMO USAR:
//   r.Use(middleware.AuthMiddleware(jwtService))
func AuthMiddleware(jwtService *jwt.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// --- 1. Extrair o token do header ---
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendAuthError(w, "header Authorization não encontrado")
				return
			}

			// --- 2. Verificar formato: "Bearer <token>" ---
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				sendAuthError(w, "formato do header inválido. Use: Bearer <token>")
				return
			}
			tokenString := parts[1]

			// --- 3. Validar o token ---
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				sendAuthError(w, "token inválido ou expirado")
				return
			}

			// --- 4. Adicionar dados no contexto ---
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			// --- 5. Chamar o próximo handler ---
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ================================================================
// MIDDLEWARE: ROLEMIDDLEWARE
// ================================================================
// Verifica se o usuário tem uma das roles permitidas.
//
// COMO USAR:
//   r.Group(func(r chi.Router) {
//       r.Use(middleware.RoleMiddleware("admin", "secretaria"))
//       r.Get("/admin", adminHandler)
//   })
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// --- 1. Extrair role do contexto ---
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				sendAuthError(w, "usuário não autenticado")
				return
			}

			// --- 2. Verificar se a role está na lista de permitidas ---
			for _, allowed := range allowedRoles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			// --- 3. Role não permitida ---
			sendAuthError(w, "permissão insuficiente para acessar este recurso")
		})
	}
}

// ================================================================
// FUNÇÃO AUXILIAR: SEND_AUTH_ERROR
// ================================================================
// Envia erro de autenticação em JSON
func sendAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}