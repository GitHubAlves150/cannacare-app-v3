// ================================================================
// PACOTE MIDDLEWARE - PERMISSIONS
// ================================================================
// Sistema de permissões baseado em roles.
//
// PERFIS E PERMISSÕES:
//   admin        - Acesso total ao sistema
//   coordenacao  - Gerenciar relatórios e aprovações
//   secretaria   - Cadastrar pacientes e documentos
//   acolhimento  - Realizar anamnese e acompanhamento
//   farmacia     - Gerenciar estoque e pedidos
//   paciente     - Acesso ao portal do paciente
//   cuidador     - Acesso ao portal em nome do paciente
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
// MAPA DE PERMISSÕES
// ================================================================
// Define quais ações cada role pode executar
var RolePermissions = map[string][]string{
	"admin": {
		"*", // Acesso total
	},
	"coordenacao": {
		"patients:read", "patients:write", "patients:approve",
		"doctors:read", "doctors:write",
		"prescriptions:read", "prescriptions:write",
		"orders:read", "orders:write",
		"financial:read", "financial:write",
		"stock:read", "stock:write",
		"reports:read",
		"dashboard:read",
	},
	"secretaria": {
		"patients:read", "patients:write",
		"doctors:read", "doctors:write",
		"prescriptions:read",
		"documents:upload", "documents:read",
		"payments:read", "payments:write",
	},
	"acolhimento": {
		"patients:read",
		"anamnesis:read", "anamnesis:write",
		"patients:update",
	},
	"farmacia": {
		"orders:read", "orders:write",
		"stock:read", "stock:write",
		"prescriptions:read",
	},
	"paciente": {
		"portal:read",
		"orders:read",
		"prescriptions:read",
		"payments:read",
	},
	"cuidador": {
		"portal:read",
		"patients:read",
		"orders:read",
		"prescriptions:read",
	},
}

// ================================================================
// FUNÇÃO CHECKPERMISSION()
// ================================================================
// Verifica se uma role tem permissão para executar uma ação
func CheckPermission(role, permission string) bool {
	// Admin tem acesso total
	if role == "admin" {
		return true
	}

	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// ================================================================
// MIDDLEWARE: PERMISSIONMIDDLEWARE
// ================================================================
// Verifica se o usuário tem uma permissão específica
func PermissionMiddleware(requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				utils.SendError(w, http.StatusUnauthorized, "usuário não autenticado")
				return
			}

			if !CheckPermission(role, requiredPermission) {
				utils.SendError(w, http.StatusForbidden, "permissão insuficiente para esta ação")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ================================================================
// MIDDLEWARE: OPTIONALAUTH (para rotas que podem ser públicas)
// ================================================================
func OptionalAuthMiddleware(jwtService *jwt.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					claims, err := jwtService.ValidateToken(parts[1])
					if err == nil {
						ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
						ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
						ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
						ctx = context.WithValue(ctx, UserNameKey, claims.Name)
						r = r.WithContext(ctx)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
