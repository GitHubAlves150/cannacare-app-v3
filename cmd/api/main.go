// ================================================================
// PACOTE MAIN
// ================================================================
// Ponto de entrada da aplicação CannaCare API.
// Inicializa todas as dependências e configura o servidor HTTP.
// ================================================================

package main

import (
	"cannacare-backend/internal/config"
	"cannacare-backend/internal/database"
	"cannacare-backend/internal/handlers"
	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"
	"cannacare-backend/pkg/jwt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware" // ✅ CORRETO	"github.com/go-chi/cors"
	"github.com/go-chi/cors"
)

// ================================================================
// FUNÇÃO MAIN()
// ================================================================
func main() {
	// ============================================================
	// PASSO 1: Carregar configurações
	// ============================================================
	log.Println("📋 Carregando configurações...")
	cfg := config.Load()

	// ============================================================
	// PASSO 2: Conectar ao banco de dados
	// ============================================================
	log.Println("🔌 Conectando ao banco de dados...")
	if err := database.Connect(cfg); err != nil {
		log.Fatal("❌ Falha ao conectar ao banco:", err)
	}
	defer database.Close()

	// ============================================================
	// PASSO 3: Rodar migrações
	// ============================================================
	log.Println("🔄 Rodando migrações...")
	if err := database.Migrate(); err != nil {
		log.Fatal("❌ Falha ao rodar migrações:", err)
	}

	// ============================================================
	// PASSO 4: Inicializar serviços
	// ============================================================
	log.Println("🔐 Inicializando serviços...")

	// Serviço JWT
	jwtService := jwt.NewJWTService(cfg.JWTSecret, cfg.JWTExpiresIn)

	// Serviço de Autenticação
	authService := services.NewAuthService(database.DB, jwtService)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)

	// ============================================================
	// PASSO 5: Configurar rotas com Chi
	// ============================================================
	log.Println("🛣️ Configurando rotas...")
	r := chi.NewRouter()

	// Middlewares globais
	r.Use(chimiddleware.Logger)    // Log de requisições
	r.Use(chimiddleware.Recoverer) // Recupera de panics
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ============================================================
	// ROTAS PÚBLICAS (sem autenticação)
	// ============================================================

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.DB.Exec("SELECT 1").Error; err != nil {
			utils.SendError(w, http.StatusServiceUnavailable, "Banco de dados indisponível")
			return
		}
		utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
			"status":   "ok",
			"database": "connected",
			"message":  "CannaCare API está funcionando!",
			"version":  "1.0.0",
			"etapa":    "3 - Autenticação",
		})
	})

	// Página inicial
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
			"message": "🌿 Bem-vindo à API CannaCare!",
			"version": "1.0.0",
			"etapa":   "3 - Autenticação",
			"endpoints": map[string]string{
				"GET  /health":            "Verifica o status do sistema",
				"POST /api/auth/register": "Registrar novo usuário",
				"POST /api/auth/login":    "Login e obter token JWT",
				"GET  /api/protected":     "Rota protegida (requer token)",
				"GET  /api/admin":         "Rota admin (requer role admin)",
			},
		})
	})

	// ============================================================
	// ROTAS DE AUTENTICAÇÃO (públicas)
	// ============================================================
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// ============================================================
	// ROTAS PROTEGIDAS (com autenticação)
	// ============================================================
	r.Group(func(r chi.Router) {
		// Aplicar middleware de autenticação em todas as rotas deste grupo
		r.Use(middleware.AuthMiddleware(jwtService))

		// Rota de teste protegida
		r.Get("/api/protected", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(middleware.UserIDKey).(string)
			role := r.Context().Value(middleware.UserRoleKey).(string)
			utils.SendSuccess(w, http.StatusOK, map[string]string{
				"message": "✅ Você está autenticado!",
				"user_id": userID,
				"role":    role,
			})
		})

		// Rota apenas para administradores
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin"))
			r.Get("/api/admin", func(w http.ResponseWriter, r *http.Request) {
				utils.SendSuccess(w, http.StatusOK, map[string]string{
					"message": "✅ Área restrita para administradores!",
				})
			})
		})
	})

	// ============================================================
	// PASSO 6: Iniciar servidor
	// ============================================================
	addr := ":" + cfg.ServerPort

	log.Println("")
	log.Println("🚀 Servidor iniciado em http://localhost" + addr)
	log.Println("📊 Health check:    http://localhost" + addr + "/health")
	log.Println("📝 Register:        POST http://localhost" + addr + "/api/auth/register")
	log.Println("🔐 Login:           POST http://localhost" + addr + "/api/auth/login")
	log.Println("🔒 Protected:       GET  http://localhost" + addr + "/api/protected (requer token)")
	log.Println("🔑 Admin:           GET  http://localhost" + addr + "/api/admin (requer role admin)")
	log.Println("")
	log.Println("💡 Pressione CTRL+C para encerrar")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal("❌ Falha ao iniciar servidor:", err)
	}
}
