package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// ===========================================================================
// Pacote config
// ===========================================================================
// Responsavel por carregar e gerenciar todas as configurações do sistema
// a partir de variaveis de ambiente.
//
// Estrutura
// 1. Define a struct Config com todos os campos necessários
// 2. Função Load() lê o arquivo .env e retorna um Config preenchido
// 3. Funcção auxiliar getenv() busca variável ou usa valor padrão
//
// COMO USAR:
// cfg := config.Load()
// fmt.Println(cfg.DBHost) // "localhost"
// ===========================================================================

// Godotenv - Carrega variáveisd do arquivo .env
// Torna mais fácil gerencial configurações por ambiente

// ===========================================================================
// STRUCT CONFIG
// ===========================================================================
// Agrupa todas as configurações do sistema em um único lugar
// Facilita a passagem de configurações entre pacotes
type Config struct {
	// ===banco de dados===
	// Configurações para conectar ao banco de dados Postgres
	DBHost     string // Endereço do banco (localhost, IP, ou nome do container)
	DBPort     string // Porta (padrão 5432)
	DBUser     string // Usuário do banco
	DBPassword string // Senha do banco
	DBName     string // Nome do banco (ex: cannacare)
	DBSSLMode  string // Modo SSL (disable, require, verify-ca, verify-full)

	// ==== SERVIDOR HTTP ====
	ServerPort string // Porta onde a API vai escutar (ex:8080)

	// ==== AMBIENTE ====
	Env string // development, production, test
}

// ===========================================================================
// FUNCAO LOAD()
// ===========================================================================
// Carrega as variaveis de ambiente e retorna um config preenchido
//
// FLUXO:
// 1. Tenta carregar o arquivo .env
// 2. Se não encontrar, usavariaveis de ambiente do sisitema
// 3. Para cada campo, busca a variável ou usa um valor padrão
//
// RETORNO; *Config(ponteiro para a struct com valores preenchidos)
// ===========================================================================
func Load() *Config {
	// Tenta carregar o arquivo .env do diretório atual
	// Se não existir, apenas loga um aviso (não para a execução)
	err := godotenv.Load()
	if err != nil {

		// Em desenvolvimento pode ser útil ter um .env
		// Em produção, usa-se variaveis de ambiente do sistema (Ex: AWS, Heroku)
		log.Println("Aviso: Arquivo .env não encontrado..")
		log.Println(" Usando variaveis de ambiente do sistema (ou valores padrão)")
	}

	// Retorna um ponteiro para a struct  Config com todos os campos preenchidos
	return &Config{
		//===BANCO DE DADOS===
		// Busca a variavel de ambiente DB_HOST
		// Se não existir, usa "localhost" como padrão
		DBHost: getEnv("DB_HOST", "localhost"),
		// Busca o DB_PORT, padrão "5432" (porta padrão do PostgresSQL)
		DBPort: getEnv("DB_Port", "5433"),
		// Busca DB_USER, padrão "postgres" (usuário padrão do PostgreSQL)
		DBUser: getEnv("DB_USER", "postgres"),
		// Busca DB_PASSWORD, padrão "postgres"
		// ⚠️ Em produção, NUNCA use senha padrão!
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		// Busca DB_NAME, padrão "cannacare"
		DBName: getEnv("DB_NAME", "cannacare"),
		// Busca DB_SSLMODE, padrão "disable" (desativado em desenvolvimento)
		// Em produção, usar "require" para conexão segura
		DBSSLMode: getEnv("DB_SSLMODE", "disable"),
		// Busca SERVER_PORT, padrão "8080"
		ServerPort: getEnv("SERVER_PORT", "8080"),
		// Busca ENV, padrão "development"
		Env: getEnv("ENV", "development"),
	}

}

// ===========================================================================
// FUNCAO AUXILIAR getEnv()
// ===========================================================================
// Busca a variavel de ambiente pelo nome (key)
// Se não existir, retorna um valor padrão (defaultValue)
// 
// PORQEU USAR?
// 1. Evita que o sisitema quebre se uma variavel nçao estiver definida
// 2. Facilita o desenvolvimento local sem precisar definir todas as variaveis
// 3. Permite definir valores padrão seguros para produção
//
// Exemplo:
// host := getEnv("DB_HOST", "localhost")
// Se DB_HOST estiver definido, usa o valor, se não, usa "localhost"
// ===========================================================================
func getEnv(key, defaultvalue string)string{
	// Os.Getenv busca a variavel de ambiente pelo nome (key)
	if value := os.Getenv(key); value!= ""{
		// Se encontrou e não está vazia então retorna o valor
		return value
	}
	//Se não encontrou, retorna o valor padrão
	return defaultvalue
}



