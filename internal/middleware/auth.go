// ================================================================
// CANNACARE - MIDDLEWARE DE AUTENTICAÇÃO
// ================================================================
// Middleware que valida o JWT em cada requisição e extrai as informações.
//
// FLUXO DE AUTENTICAÇÃO:
//   1. O cliente envia uma requisição com o header Authorization: Bearer <token>
//   2. O middleware valida o token (assinatura, expiração, etc)
//   3. Se válido, extrai as informações do token
//   4. As informações são adicionadas ao contexto da requisição
//   5. Os handlers podem acessar essas informações via Context
//
// O QUE É ADICIONADO AO CONTEXTO:
//   - user_id: ID do usuário
//   - association_id: ID da associação ← ESSENCIAL PARA MULTI-TENANCY!
//   - user_email: Email do usuário
//   - user_role: Papel/função do usuário
//   - user_name: Nome do usuário
//
// POR QUE USAR CONTEXT?
//   - O Context é uma forma de passar informações entre middlewares e handlers
//   - Evita que tenhamos que passar parâmetros manualmente
//   - As informações ficam disponíveis em toda a cadeia de handlers
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
// Chaves usadas para armazenar valores no Context.
// Usamos um tipo customizado (ContextKey) para evitar colisões.
type ContextKey string

// Constantes para as chaves do Context
const (
	// UserIDKey - Chave para o ID do usuário
	UserIDKey ContextKey = "user_id"

	// AssociationIDKey - Chave para o ID da associação
	// ⚠️ ESSENCIAL PARA MULTI-TENANCY!
	// Este valor é usado em TODAS as queries para filtrar dados
	AssociationIDKey ContextKey = "association_id"

	// UserEmailKey - Chave para o email do usuário
	UserEmailKey ContextKey = "user_email"

	// UserRoleKey - Chave para o papel/função do usuário
	UserRoleKey ContextKey = "user_role"

	// UserNameKey - Chave para o nome do usuário
	UserNameKey ContextKey = "user_name"
)

// ================================================================
// MIDDLEWARE: AUTHMIDDLEWARE
// ================================================================
// Middleware que valida o JWT e adiciona as informações ao contexto.
//
// COMO USAR:
//   r.Use(middleware.AuthMiddleware(jwtService))
//
// FLUXO:
//   1. Extrai o token do header Authorization
//   2. Valida o token com o JWTService
//   3. Extrai as claims (informações)
//   4. Adiciona as informações ao Context
//   5. Chama o próximo handler
func AuthMiddleware(jwtService *jwt.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// --- PASSO 1: Extrair o token do header ---
			// O header deve ser: Authorization: Bearer <token>
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.SendError(w, http.StatusUnauthorized, "header Authorization não encontrado")
				return
			}

			// --- PASSO 2: Verificar o formato ---
			// Deve ser "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.SendError(w, http.StatusUnauthorized, "formato inválido. Use: Bearer <token>")
				return
			}
			tokenString := parts[1]

			// --- PASSO 3: Validar o token ---
			// O JWTService verifica assinatura, expiração, etc
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				utils.SendError(w, http.StatusUnauthorized, "token inválido ou expirado")
				return
			}

			// --- PASSO 4: Adicionar informações ao Context ---
			// O Context é passado para todos os handlers subsequentes
			// O association_id é ESSENCIAL para filtrar dados!
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, AssociationIDKey, claims.AssociationID) // ← MULTI-TENANCY!
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, UserNameKey, claims.Name)

			// --- PASSO 5: Chamar o próximo handler ---
			// O próximo handler (ex: PatientHandler) pode acessar as informações
			// usando r.Context().Value(middleware.UserIDKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ================================================================
// MIDDLEWARE: ROLEMIDDLEWARE
// ================================================================
// Middleware que verifica se o usuário tem uma das roles permitidas.
//
// COMO USAR:
//   r.Use(middleware.RoleMiddleware("admin", "coordenacao"))
//
// FLUXO:
//   1. Extrai a role do Context (definida pelo AuthMiddleware)
//   2. Verifica se a role está na lista de permitidas
//   3. Se não estiver, retorna erro 403 (Forbidden)
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extrai a role do Context
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				utils.SendError(w, http.StatusUnauthorized, "usuário não autenticado")
				return
			}

			// Verifica se a role está na lista de permitidas
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