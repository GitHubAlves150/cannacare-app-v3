package database

// ================================================================
// PACOTE DATABASE
// ================================================================
// Gerencia a conexão com o banco de dados postrgres usando Gorm
//
// RESPONSABILIDADES:
// 1. Estabelecer a conexão com o banco
// 2. Manter a conexão com o banco
// 3. Fornecer uma instância global do *gorm.DB
//
//PADRÃO UTILIZADO: singleton
// -A Conexão é criada uma única vez (no Connect)
// -Todas as operações usam a mesma instância (DB global)
// -Evita múltiplas conexões desnecessárias
// ================================================================

import (
	"fmt"
	"log"

	// Importa o pacote config para acessar as configurações
	"cannacare-backend/internal/config"

	// Driver PostgreSQL para o GORM
	// O underline (_) indica que só queremos o efeito colateral (init)
	// O GORM usa o driver internamente, não precisamos chamar diretamente
	"gorm.io/driver/postgres"

	// GORM - ORM principal
	"gorm.io/gorm"

	// Logger do GORM - controla o nível de detalhe dos logs
	"gorm.io/gorm/logger"
)

// ================================================================
// VARIÁVEL GLOBAL DB
// ================================================================
// Armazena a instãncia única do *gorm.DB
// È exportada (maiúscula) para ser acessada por outros pacotes
//
//COMO USAR:
// //Em handler
// db:= database.GetDB()
// var patiente models.Patient
// db.First(&patient, "id = ?", id)

var DB *gorm.DB

// ================================================================
// FUNÇÃO CONNECT()
// ================================================================
// Estabelece a conexão com o banco de dados PostgreSQL
//
// PARAMETROS:
// cfg *config.Config -Configurações carregadas do .env
//
// RETORNO:
// erro - se houver erro na conexão, retonra o err
//
// FLUXO:
// 1. Monta a DSN (Data Source Name) - string de conexão
// 2. Tenta abrir a conexão com o GORM
// 3. Configura o logger para mostrar todas as queries
// 4. Armazena na variável global DB
//
// EXEMPLO DE DSN:
//
//	"host=localhost user=postgres password=postgres dbname=cannacare port=5432 sslmode=disable"
//
// ================================================================
func Connect(cfg *config.Config) error {
	//===PASSO 1: Montar a dsn ====
	//DSN = Data Source Name = string de conexão
	// Formato: "host=... user=... password=... dbname=... port=... sslmode=..."
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	//log para debug - mostra qual banco está conectado
	//Nunca logue a senha em produção
	log.Printf("..Conectado ao banco %s: %s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	// PASSO 2: Abrir conexão com o GORM
	// Postgres.Open() cria um driver PostgreSQL
	// &gorm.Config{} define a configurações do GORM
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: Define o nível de detalhe dos logs
		// logger.Info = Mostra todas as queries SQL executadas
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		// se não conseguiu conectar, retorna erro com contexto
		return fmt.Errorf("Falha ao conectar ao banco: %s", err)
	}

	// passo 3: verificar se a conecxão está ativa
	// O GORM pode conectar, mas o banco pode estar inativo
	// ping() verifica se a conexão funciona

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("Falah ao obter conexão: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("Banco não respondeu ao ping: %w", err)
	}

	log.Println("Conexão com o banco estabelecida com sucesso")
	return nil
}

// ================================================================
// FUNÇÃO GETDB()
// ================================================================
// Retorna a instância global do *gorm.DB para ser usada em outras partes
// do sistema (handler, services, repositories)
//
// db := dstabase.GetDB()
// var patients []models.Patients
// db.Find(&patients)
// ================================================================
func GetDB() *gorm.DB {
	return DB
}

// ================================================================
// FUNÇÃO CLOSE()
// ================================================================
// Fecha aconexão com o banco de dados
// Deve ser chamada quando a apllicação for encerrada (defer)
//
// USO:
// defer database.Close()
//
// Retorno:
// erro - Se tiver erro ao fechar, retorna o erro
// ================================================================
func Close() error {
	// Obté a conexão SQL subjacente
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	//Fecha a conexão
	log.Println("Fechando conexão com o banco...")
	return sqlDB.Close()
}
