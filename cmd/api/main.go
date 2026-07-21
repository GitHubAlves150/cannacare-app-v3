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
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// ============================================================
	// PASSO 1: Carregar configurações
	// ============================================================
	log.Println("📋 Carregando configurações...")
	cfg := config.Load()

	// ============================================================
	// PASSO 2: Conectar ao banco
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

	// JWT
	jwtService := jwt.NewJWTService(cfg.JWTSecret, cfg.JWTExpiresIn)

	// Serviços
	authService := services.NewAuthService(database.DB, jwtService)
	doctorService := services.NewDoctorService(database.DB)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	doctorHandler := handlers.NewDoctorHandler(doctorService)

	// ============================================================
	// PASSO 5: Configurar rotas
	// ============================================================
	log.Println("🛣️ Configurando rotas...")
	r := chi.NewRouter()

	// Middlewares globais
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ============================================================
	// ROTAS PÚBLICAS
	// ============================================================
	r.Get("/health", healthCheckHandler)
	r.Get("/", welcomeHandler)

	// ============================================================
	// ROTAS DE AUTENTICAÇÃO
	// ============================================================
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// ============================================================
	// ROTAS PROTEGIDAS (com autenticação)
	// ============================================================
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtService))

		// --- Rota de teste ---
		r.Get("/api/protected", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(middleware.UserIDKey).(string)
			role := r.Context().Value(middleware.UserRoleKey).(string)
			utils.SendSuccess(w, http.StatusOK, map[string]string{
				"message": "✅ Você está autenticado!",
				"user_id": userID,
				"role":    role,
			})
		})

		// --- Rotas de Médicos (requer role admin ou secretaria) ---
		r.Group(func(r chi.Router) {
			// Permitir admin e secretaria
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao"))

			r.Post("/api/doctors", doctorHandler.Create)
			r.Get("/api/doctors", doctorHandler.List)
			r.Get("/api/doctors/top", doctorHandler.GetTopDoctors)
			r.Get("/api/doctors/{id}", doctorHandler.GetByID)
			r.Put("/api/doctors/{id}", doctorHandler.Update)
			r.Delete("/api/doctors/{id}", doctorHandler.Delete)
		})

		// --- Rota Admin ---
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
	log.Println("📊 Health check:        http://localhost" + addr + "/health")
	log.Println("📝 Register:            POST http://localhost" + addr + "/api/auth/register")
	log.Println("🔐 Login:               POST http://localhost" + addr + "/api/auth/login")
	log.Println("👨‍⚕️ Doctors:            http://localhost" + addr + "/api/doctors")
	log.Println("🔒 Protected:           GET  http://localhost" + addr + "/api/protected")
	log.Println("🔑 Admin:               GET  http://localhost" + addr + "/api/admin")
	log.Println("")
	log.Println("💡 Pressione CTRL+C para encerrar")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal("❌ Falha ao iniciar servidor:", err)
	}
}

// ================================================================
// HANDLERS AUXILIARES
// ================================================================

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := database.DB.Exec("SELECT 1").Error; err != nil {
		utils.SendError(w, http.StatusServiceUnavailable, "Banco de dados indisponível")
		return
	}
	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"status":   "ok",
		"database": "connected",
		"message":  "CannaCare API está funcionando!",
		"version":  "1.0.0",
		"etapa":    "4 - Médicos",
	})
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "🌿 Bem-vindo à API CannaCare!",
		"version": "1.0.0",
		"etapa":   "4 - Médicos",
		"endpoints": map[string]string{
			"GET  /health":             "Verifica o status do sistema",
			"POST /api/auth/register":  "Registrar novo usuário",
			"POST /api/auth/login":     "Login e obter token JWT",
			"GET  /api/protected":      "Rota protegida (requer token)",
			"GET  /api/admin":          "Rota admin (requer role admin)",
			"POST /api/doctors":        "Criar médico (admin/secretaria)",
			"GET  /api/doctors":        "Listar médicos (admin/secretaria)",
			"GET  /api/doctors/{id}":   "Buscar médico por ID",
			"PUT  /api/doctors/{id}":   "Atualizar médico",
			"DELETE /api/doctors/{id}": "Remover médico",
			"GET  /api/doctors/top":    "Médicos que mais prescrevem",
		},
	})
}
