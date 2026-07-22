package main

import (
	"log"
	"net/http"

	"cannacare-backend/internal/config"
	"cannacare-backend/internal/database"
	"cannacare-backend/internal/handlers"
	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"
	"cannacare-backend/pkg/jwt"

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

	jwtService := jwt.NewJWTService(cfg.JWTSecret, cfg.JWTExpiresIn)

	authService := services.NewAuthService(database.DB, jwtService)
	doctorService := services.NewDoctorService(database.DB)
	patientService := services.NewPatientService(database.DB) // 🆕

	authHandler := handlers.NewAuthHandler(authService)
	doctorHandler := handlers.NewDoctorHandler(doctorService)
	patientHandler := handlers.NewPatientHandler(patientService) // 🆕

	// ============================================================
	// PASSO 5: Configurar rotas
	// ============================================================
	log.Println("🛣️ Configurando rotas...")
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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
	// ROTAS PROTEGIDAS
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

		// --- Rotas de Médicos ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao"))

			r.Post("/api/doctors", doctorHandler.Create)
			r.Get("/api/doctors", doctorHandler.List)
			r.Get("/api/doctors/top", doctorHandler.GetTopDoctors)
			r.Get("/api/doctors/{id}", doctorHandler.GetByID)
			r.Put("/api/doctors/{id}", doctorHandler.Update)
			r.Delete("/api/doctors/{id}", doctorHandler.Delete)
		})

		// Inicializar Stock Service
		// --- Rotas estoque ----
		stockService := services.NewStockService(database.DB)

		// Inicializar Stock Handler
		stockHandler := handlers.NewStockHandler(stockService)

		// Adicionar as rotas de estoque no grupo protegido:
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "farmacia"))

			// Rotas de Estoque
			r.Post("/api/stock/lots", stockHandler.CreateLot)
			r.Get("/api/stock/lots", stockHandler.ListLots)
			r.Get("/api/stock/lots/{id}", stockHandler.GetLotByID)
			r.Post("/api/stock/adjust", stockHandler.AdjustStock)
			r.Get("/api/stock/movements", stockHandler.GetMovements)
			r.Get("/api/stock/expiring", stockHandler.GetExpiringLots)
			r.Get("/api/stock/low-stock", stockHandler.GetLowStock)
			r.Get("/api/stock/summary", stockHandler.GetStockSummary)
		})

		// Inicializar Product Service
		// --- Rotas de produtos ----
		productService := services.NewProductService(database.DB)

		// Inicializar Product Handler
		productHandler := handlers.NewProductHandler(productService)

		// Adicionar as rotas de produtos no grupo protegido:
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao"))

			// Rotas de Produtos
			r.Post("/api/products", productHandler.Create)
			r.Get("/api/products", productHandler.List)
			r.Get("/api/products/low-stock", productHandler.GetLowStock)
			r.Get("/api/products/stock-summary", productHandler.GetStockSummary)
			r.Get("/api/products/{id}", productHandler.GetByID)
			r.Put("/api/products/{id}", productHandler.Update)
			r.Delete("/api/products/{id}", productHandler.Delete)
		})

		// Inicializar Anamnese Service
		// ----Rotas anamenese -----
		anamneseService := services.NewAnamneseService(database.DB)

		// Inicializar Anamnese Handler
		anamneseHandler := handlers.NewAnamneseHandler(anamneseService)

		// Adicionar as rotas de anamnese no grupo protegido:
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "acolhimento"))

			// Rotas de Anamnese
			r.Post("/api/patients/{id}/anamnesis", anamneseHandler.Create)
			r.Get("/api/patients/{id}/anamnesis", anamneseHandler.GetByPatient)
			r.Get("/api/anamnesis", anamneseHandler.List)
			r.Get("/api/anamnesis/{id}", anamneseHandler.GetByID)
			r.Put("/api/anamnesis/{id}", anamneseHandler.Update)
			r.Delete("/api/anamnesis/{id}", anamneseHandler.Delete)
		})

		// -----Rotas de presciçoes/receitas médicas ----
		prescriptionService := services.NewPrescriptionService(database.DB)
		prescriptionHandler := handlers.NewPrescriptionHandler(prescriptionService)
		// Adicionar as rotas de prescrições no grupo protegido:
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "acolhimento"))

			// Rotas de Prescrições
			r.Post("/api/prescriptions", prescriptionHandler.Create)
			r.Get("/api/prescriptions", prescriptionHandler.List)
			r.Get("/api/prescriptions/expired", prescriptionHandler.GetExpired)
			r.Get("/api/prescriptions/validate/{id}", prescriptionHandler.Validate)
			r.Get("/api/prescriptions/{id}", prescriptionHandler.GetByID)
			r.Put("/api/prescriptions/{id}", prescriptionHandler.Update)
			r.Delete("/api/prescriptions/{id}", prescriptionHandler.Delete)
			r.Post("/api/prescriptions/update-status", prescriptionHandler.UpdateAllStatus)
		})

		// --- 🆕 Rotas de Pacientes ---
		// Inicializar serviços
		documentService := services.NewDocumentService(database.DB)

		// Inicializar handlers
		documentHandler := handlers.NewDocumentHandler(documentService)
		r.Group(func(r chi.Router) {
			// Acesso para admin, secretaria, coordenacao e acolhimento
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "acolhimento"))

			r.Post("/api/patients", patientHandler.Create)
			r.Get("/api/patients", patientHandler.List)
			r.Get("/api/patients/stats", patientHandler.GetStatistics)
			r.Get("/api/patients/{id}", patientHandler.GetByID)
			r.Put("/api/patients/{id}", patientHandler.Update)
			r.Patch("/api/patients/{id}/status", patientHandler.UpdateStatus)
			r.Delete("/api/patients/{id}", patientHandler.Delete)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "acolhimento"))

			r.Post("/api/patients/{id}/documents", documentHandler.Upload)
			r.Get("/api/patients/{id}/documents", documentHandler.ListByPatient)
			r.Get("/api/documents/{id}", documentHandler.GetByID)
			r.Get("/api/documents/{id}/download", documentHandler.Download)
			r.Patch("/api/documents/{id}/status", documentHandler.UpdateStatus)
			r.Delete("/api/documents/{id}", documentHandler.Delete)
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
	log.Println("👤 Patients:            http://localhost" + addr + "/api/patients")
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
		"etapa":    "5 - Pacientes",
	})
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "🌿 Bem-vindo à API CannaCare!",
		"version": "1.0.0",
		"etapa":   "5 - Pacientes",
		"endpoints": map[string]string{
			"GET  /health":                    "Verifica o status do sistema",
			"POST /api/auth/register":         "Registrar novo usuário",
			"POST /api/auth/login":            "Login e obter token JWT",
			"GET  /api/protected":             "Rota protegida (requer token)",
			"GET  /api/admin":                 "Rota admin (requer role admin)",
			"POST /api/doctors":               "Criar médico",
			"GET  /api/doctors":               "Listar médicos",
			"GET  /api/doctors/{id}":          "Buscar médico por ID",
			"PUT  /api/doctors/{id}":          "Atualizar médico",
			"DELETE /api/doctors/{id}":        "Remover médico",
			"GET  /api/doctors/top":           "Médicos que mais prescrevem",
			"POST /api/patients":              "Criar paciente",
			"GET  /api/patients":              "Listar pacientes",
			"GET  /api/patients/{id}":         "Buscar paciente por ID",
			"PUT  /api/patients/{id}":         "Atualizar paciente",
			"PATCH /api/patients/{id}/status": "Mudar status do paciente",
			"DELETE /api/patients/{id}":       "Remover paciente",
			"GET  /api/patients/stats":        "Estatísticas de pacientes",
		},
	})
}
