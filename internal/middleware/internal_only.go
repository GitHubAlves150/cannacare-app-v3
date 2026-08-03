// ================================================================
// CANNACARE - MIDDLEWARE INTERNAL ONLY
// ================================================================
// Protege endpoints que não devem ficar abertos ao público, mas
// também não fazem sentido atrás de login normal (ex: /api/auth/register,
// que cria uma associação nova do zero — hoje só usado manualmente
// pelo próprio time do CannaCare para parcerias, migrações e testes,
// já que o fluxo real de clientes passa pelo site com CNPJ validado).
//
// Exige o header:
//   X-Internal-Key: <valor de INTERNAL_API_KEY no .env>
//
// ⚠️ Isso NÃO é autenticação de usuário — é só uma chave compartilhada
// entre quem administra o sistema. Não usar para nada acessível por
// clientes finais.
// ================================================================

package middleware

import (
	"net/http"
	"os"

	"cannacare-backend/internal/utils"
)

func InternalOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedKey := os.Getenv("INTERNAL_API_KEY")

		if expectedKey == "" {
			// Sem chave configurada = endpoint fica bloqueado por padrão,
			// nunca aberto por engano.
			utils.SendError(w, http.StatusServiceUnavailable, "endpoint interno não configurado")
			return
		}

		providedKey := r.Header.Get("X-Internal-Key")
		if providedKey == "" || providedKey != expectedKey {
			utils.SendError(w, http.StatusUnauthorized, "acesso negado")
			return
		}

		next.ServeHTTP(w, r)
	})
}