package main

import (
	"log"
	"net/http"

	"cannacare-backend/internal/config"
	"cannacare-backend/internal/database"
)

func main() {
	log.Println("📋 Carregando configurações...")
	cfg := config.Load()

	log.Println("🔌 Conectando ao banco de dados...")
	if err := database.Connect(cfg); err != nil {
		log.Fatal("❌ Falha ao conectar ao banco:", err)
	}
	defer database.Close()

	// ============================================================
	// 🆕 NOVO: RODAR MIGRAÇÕES
	// ============================================================
	// Cria/atualiza as tabelas automaticamente
	// ⚠️ Em produção, usar migrações manuais
	log.Println("🔄 Rodando migrações...")
	if err := database.Migrate(); err != nil {
		log.Fatal("❌ Falha ao rodar migrações:", err)
	}

	// ============================================================
	// CONFIGURAR ROTAS
	// ============================================================
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/", welcomeHandler)

	// ============================================================
	// INICIAR SERVIDOR
	// ============================================================
	addr := ":" + cfg.ServerPort

	log.Printf("🚀 Servidor iniciado em http://localhost%s", addr)
	log.Println("📊 Health check: http://localhost" + addr + "/health")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("❌ Falha ao iniciar servidor:", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := database.DB.Exec("SELECT 1").Error; err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{
			"status": "error",
			"database": "disconnected",
			"message": "Banco de dados indisponível"
		}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"status": "ok",
		"database": "connected",
		"message": "CannaCare API está funcionando!",
		"version": "1.0.0",
		"migrations": "✅ Todas as tabelas foram criadas"
	}`))
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"message": "🌿 Bem-vindo à API CannaCare!",
		"version": "1.0.0",
		"status": "Etapa 2 concluída - Models + Migrations",
		"endpoints": {
			"GET /health": "Verifica o status do sistema e banco de dados"
		},
		"tables_created": 16
	}`))
}