// ================================================================
// PACOTE MIDDLEWARE - AUTH
// ================================================================
// Middleware para autenticação JWT e roles/permissões.
// ================================================================

package middleware

import (
	"context"
	"net/http"
	"strings"

	"cannacare-backend/internal/utils"
	"cannacare-backend/pkg/jwt"
)

// ================================================================
// CONTEXT KEYS
// ================================================================
type ContextKey string

const (
	UserIDKey    ContextKey = "user_id"
	UserEmailKey ContextKey = "user_email"
	UserRoleKey  ContextKey = "user_role"
	UserNameKey  ContextKey = "user_name"
)

// ================================================================
// MIDDLEWARE: AUTHMIDDLEWARE
// ================================================================
func AuthMiddleware(jwtService *jwt.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// --- 1. Extrair token do header ---
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.SendError(w, http.StatusUnauthorized, "header Authorization não encontrado")
				return
			}

			// --- 2. Verificar formato ---
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.SendError(w, http.StatusUnauthorized, "formato inválido. Use: Bearer <token>")
				return
			}
			tokenString := parts[1]

			// --- 3. Validar token ---
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				utils.SendError(w, http.StatusUnauthorized, "token inválido ou expirado")
				return
			}

			// --- 4. Adicionar dados no contexto ---
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, UserNameKey, claims.Name)

			// --- 5. Chamar próximo handler ---
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ================================================================
// MIDDLEWARE: ROLEMIDDLEWARE
// ================================================================
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				utils.SendError(w, http.StatusUnauthorized, "usuário não autenticado")
				return
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			utils.SendError(w, http.StatusForbidden, "permissão insuficiente")
		})
	}
}

// ================================================================
// MIDDLEWARE: GET USER FROM CONTEXT (Helper)
// ================================================================
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}

func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}
