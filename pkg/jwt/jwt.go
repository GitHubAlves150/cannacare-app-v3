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
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`   // 🆕 ADICIONAR CAMPO NAME
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ================================================================
// STRUCT JWTSERVICE
// ================================================================
type JWTService struct {
	secretKey string
	expiresIn time.Duration
}

// ================================================================
// FUNÇÃO NEWJWTSERVICE()
// ================================================================
func NewJWTService(secretKey string, expiresIn time.Duration) *JWTService {
	return &JWTService{
		secretKey: secretKey,
		expiresIn: expiresIn,
	}
}

// ================================================================
// FUNÇÃO EXPIRESIN()
// ================================================================
func (s *JWTService) ExpiresIn() time.Duration {
	return s.expiresIn
}

// ================================================================
// FUNÇÃO GENERATETOKEN()
// ================================================================
func (s *JWTService) GenerateToken(userID uuid.UUID, email, name, role string) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		Email:  email,
		Name:   name,   // 🆕 ADICIONAR NAME
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// ================================================================
// FUNÇÃO VALIDATETOKEN()
// ================================================================
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}