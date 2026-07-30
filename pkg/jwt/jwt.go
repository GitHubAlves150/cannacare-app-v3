// ================================================================
// CANNACARE - JWT (JSON Web Token)
// ================================================================
// Gerencia a criação e validação de tokens JWT para autenticação.
//
// O JWT é um token que contém informações do usuário e é assinado
// digitalmente. Ele é enviado pelo cliente em todas as requisições
// para provar que está autenticado.
//
// ESTRUTURA DO JWT:
//   Header: { "alg": "HS256", "typ": "JWT" }
//   Payload: { "user_id": "...", "association_id": "...", ... }
//   Signature: assinatura digital
//
// O QUE O JWT CONTÉM (Claims):
//   - user_id: ID do usuário
//   - association_id: ID da associação ← ESSENCIAL PARA MULTI-TENANCY!
//   - email: Email do usuário
//   - name: Nome do usuário
//   - role: Papel/função do usuário
//   - exp: Data de expiração
//
// POR QUE O ASSOCIATION_ID VAI NO JWT?
//   - O token é enviado pelo cliente em TODAS as requisições
//   - O middleware extrai o association_id do token
//   - Todas as queries filtram por association_id
//   - Isso garante que o usuário SÓ veja dados da sua associação
// ================================================================

package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ================================================================
// STRUCT CLAIMS
// ================================================================
// Claims contém TODAS as informações que serão armazenadas no JWT.
// Quando o usuário faz login, geramos um token com estas informações.
// Quando o usuário faz uma requisição, extraímos estas informações do token.
type Claims struct {
	// UserID - ID do usuário (identifica quem está logado)
	UserID string `json:"user_id"`

	// AssociationID - ID da associação (identifica qual cliente)
	// ⚠️ ESSENCIAL PARA MULTI-TENANCY!
	// Este campo é usado para filtrar TODAS as queries no banco.
	// Sem ele, um usuário de uma associação poderia ver dados de outra.
	AssociationID string `json:"association_id"`

	// Email - Email do usuário (para identificação)
	Email string `json:"email"`

	// Name - Nome do usuário (para exibição)
	Name string `json:"name"`

	// Role - Papel/função do usuário (admin, secretaria, etc)
	Role string `json:"role"`

	// RegisteredClaims - Campos padrão do JWT (expiração, emissão, etc)
	jwt.RegisteredClaims
}

// ================================================================
// STRUCT JWTSERVICE
// ================================================================
// Service que gerencia a criação e validação de tokens.
type JWTService struct {
	secretKey string          // Chave secreta para assinar os tokens
	expiresIn time.Duration   // Tempo de expiração (ex: 24h)
}

// ================================================================
// FUNÇÃO NEWJWTSERVICE()
// ================================================================
// Cria uma nova instância do serviço JWT.
func NewJWTService(secretKey string, expiresIn time.Duration) *JWTService {
	return &JWTService{
		secretKey: secretKey,
		expiresIn: expiresIn,
	}
}

// ================================================================
// FUNÇÃO GENERATETOKEN()
// ================================================================
// Gera um novo token JWT para o usuário.
//
// PARÂMETROS:
//   - userID: ID do usuário
//   - associationID: ID da associação (MULTI-TENANCY!)
//   - email: Email do usuário
//   - name: Nome do usuário
//   - role: Papel/função do usuário
//
// RETORNO:
//   - string: O token JWT (ex: "eyJhbGciOiJIUzI1NiIs...")
//   - error: Erro se houver falha na geração
//
// EXEMPLO DE TOKEN GERADO:
//   Header: { "alg": "HS256", "typ": "JWT" }
//   Payload: {
//     "user_id": "abc-123",
//     "association_id": "550e8400-...", ← ESSENCIAL!
//     "email": "admin@associacao.com",
//     "name": "Administrador",
//     "role": "admin",
//     "exp": 1735689600
//   }
func (s *JWTService) GenerateToken(userID uuid.UUID, associationID uuid.UUID, email, name, role string) (string, error) {
	// Cria as claims (informações do token)
	claims := Claims{
		UserID:        userID.String(),
		AssociationID: associationID.String(), // ← ADICIONADO PARA MULTI-TENANCY!
		Email:         email,
		Name:          name,
		Role:          role,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt: Quando o token expira
			// O usuário precisa fazer login novamente após este período
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiresIn)),

			// IssuedAt: Quando o token foi emitido
			IssuedAt: jwt.NewNumericDate(time.Now()),

			// NotBefore: A partir de quando o token é válido
			// Definimos como agora para ser válido imediatamente
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Cria o token com o algoritmo HS256 (HMAC com SHA-256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Assina o token com a chave secreta e retorna
	return token.SignedString([]byte(s.secretKey))
}

// ================================================================
// FUNÇÃO VALIDATETOKEN()
// ================================================================
// Valida um token JWT e extrai as claims (informações).
//
// PARÂMETROS:
//   - tokenString: O token JWT (ex: "eyJhbGciOiJIUzI1NiIs...")
//
// RETORNO:
//   - *Claims: As informações do token (user_id, association_id, etc)
//   - error: Erro se o token for inválido ou expirado
//
// COMO É USADO:
//   O middleware chama esta função para validar o token de cada requisição.
//   Se for válido, as claims são adicionadas ao contexto da requisição.
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse do token com as claims
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verifica se o método de assinatura é HMAC (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		// Retorna a chave secreta para verificar a assinatura
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// Se o token for válido, retorna as claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}

// ================================================================
// FUNÇÃO EXPIRESIN()
// ================================================================
// Retorna o tempo de expiração do token.
// Usado para informar o cliente sobre o tempo de vida do token.
func (s *JWTService) ExpiresIn() time.Duration {
	return s.expiresIn
}