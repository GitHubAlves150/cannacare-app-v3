// ================================================================
// PACOTE MAIN
// ================================================================
// Ponto de entrada da aplicação.
// É aqui que o servidor é inicializado e todas as peças são conectadas.
//
// FLUXO DE INICIALIZAÇÃO:
// 1. Carregar configurações (.env)
// 2. Conectar ao banco de dados
// 3. Verificar a conexão com ping
// 4. Iniciar servidor HTTP
// 5. Aguardar requisições
//
// OBSERVAÇÕES:
// - Usa o pacote net/http padrão do Go (sem frameworks ainda)
// - Próximas etapas substituirão por Chi (framework mais completo)
// ================================================================

package main

import (
	"log"
	"net/http"

	// Importa os pacotes internos do projeto
	"cannacare-backend/internal/config"
	"cannacare-backend/internal/database"
)

// ================================================================
// FUNÇÃO MAIN()
// ================================================================
// Ponto de entrada executado quando o programa inicia
//
// ORDEM DE EXECUÇÃO:
// 1. Carregar configurações
// 2. Conectar ao banco
// 3. Verificar conexão
// 4. Configurar rotas
// 5. Iniciar servidor
// ================================================================
func main() {
	// ============================================================
	// PASSO 1: Carregar Configurações
	// ============================================================
	// Carrega as variáveis de ambiente do arquivo .env
	// Cria uma struct Config com todos os valores
	log.Println("📋 Carregando configurações...")
	cfg := config.Load()

	// ============================================================
	// PASSO 2: Conectar ao Banco de Dados
	// ============================================================
	// Estabelece a conexão com PostgreSQL usando GORM
	// Se falhar, a aplicação não pode continuar
	log.Println("🔌 Conectando ao banco de dados...")
	if err := database.Connect(cfg); err != nil {
		// log.Fatal() exibe o erro e encerra o programa (código 1)
		// Não continuamos sem banco de dados
		log.Fatal("❌ Falha ao conectar ao banco:", err)
	}

	// Garantir que a conexão seja fechada ao final da execução
	// O "defer" executa a função quando a função atual (main) terminar
	// Importante para liberar recursos do banco
	defer func() {
		log.Println("🔄 Fechando conexões...")
		if err := database.Close(); err != nil {
			log.Println("⚠️ Erro ao fechar conexão:", err)
		}
	}()

	// ============================================================
	// PASSO 3: Verificar Conexão
	// ============================================================
	// O Connect já faz Ping, mas reforçamos aqui para garantir
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Fatal("❌ Falha ao obter conexão:", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("❌ Banco não respondeu ao ping:", err)
	}
	log.Println("✅ Banco de dados está respondendo!")

	// ============================================================
	// PASSO 4: Configurar Rotas
	// ============================================================
	// Usando o pacote net/http padrão (simples)
	// Nas próximas etapas, substituiremos por Chi/gin

	// Rota /health - Verifica se a API e o banco estão funcionando
	// Útil para monitoramento e health checks em containers (Kubernetes, Docker)
	http.HandleFunc("/health", healthCheckHandler)

	// Rota / - Mensagem de boas-vindas
	http.HandleFunc("/", welcomeHandler)

	// ============================================================
	// PASSO 5: Iniciar Servidor
	// ============================================================
	addr := ":" + cfg.ServerPort

	log.Printf("🚀 Servidor iniciado em http://localhost%s", addr)
	log.Println("📊 Health check: http://localhost" + addr + "/health")
	log.Println("📋 API base: http://localhost" + addr + "/")
	log.Println("")
	log.Println("💡 Pressione CTRL+C para encerrar")

	// Inicia o servidor HTTP
	// ListenAndServe bloqueia a execução (fica ouvindo requisições)
	// Se falhar, encerra a aplicação
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("❌ Falha ao iniciar servidor:", err)
	}
}

// ================================================================
// FUNÇÃO HEALTHCHECKHANDLER
// ================================================================
// Handler para o endpoint /health
// Retorna o status da API e da conexão com o banco
//
// USO:
//   GET /health
//
// RESPOSTA (JSON):
//   {
//     "status": "ok",
//     "database": "connected",
//     "message": "CannaCare API está funcionando!"
//   }
//
// STATUS HTTP:
//   200 OK - Tudo funcionando
//   503 Service Unavailable - Se o banco não responder
// ================================================================
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica se o banco está respondendo
	if err := database.DB.Exec("SELECT 1").Error; err != nil {
		// Se o banco não responder, retorna erro 503
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{
			"status": "error",
			"database": "disconnected",
			"message": "Banco de dados indisponível"
		}`))
		return
	}

	// Tudo funcionando -> 200 OK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"status": "ok",
		"database": "connected",
		"message": "CannaCare API está funcionando!",
		"version": "1.0.0"
	}`))
}

// ================================================================
// FUNÇÃO WELCOMEHANDLER
// ================================================================
// Handler para a rota raiz (/)
// Mostra uma mensagem de boas-vindas e lista endpoints disponíveis
//
// USO:
//   GET /
//
// RESPOSTA (JSON):
//   {
//     "message": "Bem-vindo à API CannaCare!",
//     "endpoints": {
//       "/health": "Verifica o status do sistema"
//     }
//   }
// ================================================================
func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"message": "🌿 Bem-vindo à API CannaCare!",
		"version": "1.0.0",
		"endpoints": {
			"GET /health": "Verifica o status do sistema e banco de dados"
		},
		"status": "Desenvolvimento ativo - Etapa 1 concluída"
	}`))
}