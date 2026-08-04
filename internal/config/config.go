// ================================================================
// CANNACARE - CONFIG (CORRIGIDO)
// ================================================================
// ⚠️ CORRIGIDO: JWT_SECRET não tem mais valor padrão. Antes, se a
// variável de ambiente não fosse definida, o servidor subia
// silenciosamente com a chave "cannacare-super-secret-key-2026"
// hardcoded no código-fonte (e visível pra qualquer um com acesso
// ao repositório) — isso permitiria forjar tokens JWT válidos.
// Agora o servidor RECUSA subir sem uma chave real e forte definida.
// ================================================================

package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// === BANCO DE DADOS ===
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// === SERVIDOR ===
	ServerPort string

	// === JWT ===
	JWTSecret    string
	JWTExpiresIn time.Duration

	// === AMBIENTE ===
	Env string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Aviso: Arquivo .env não encontrado, usando variáveis de ambiente")
	}

	jwtSecret := requireStrongSecret("JWT_SECRET")

	return &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "cannacare"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// JWT
		JWTSecret:    jwtSecret,
		JWTExpiresIn: getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour),

		// Environment
		Env: getEnv("ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}
	return defaultValue
}

// ================================================================
// requireStrongSecret - exige a variável definida E forte o bastante
// ================================================================
// Derruba o servidor na inicialização (não numa requisição aleatória
// depois) se a chave estiver ausente ou for fraca demais pra assinar
// tokens JWT com segurança.
func requireStrongSecret(key string) string {
	value := os.Getenv(key)

	if value == "" {
		log.Fatalf("❌ %s não está definido. Configure uma chave forte no .env antes de subir o servidor. "+
			"Gere uma com: openssl rand -base64 48", key)
	}

	if len(value) < 32 {
		log.Fatalf("❌ %s muito curto (%d caracteres). Use pelo menos 32 caracteres. "+
			"Gere uma com: openssl rand -base64 48", key, len(value))
	}

	// Bloqueia explicitamente o valor antigo que ficava hardcoded no
	// código, pra ninguém copiar e colar ele "só pra testar" e esquecer.
	if value == "cannacare-super-secret-key-2026" || value == "cannacare-super-secret-key-2026-change-in-production" {
		log.Fatalf("❌ %s ainda está com o valor de exemplo antigo. Gere uma chave nova com: openssl rand -base64 48", key)
	}

	return value
}