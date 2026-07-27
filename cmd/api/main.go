package main

import (
	"log"
	"net/http"

	"cannacare-backend/internal/config"
	"cannacare-backend/internal/database"
	"cannacare-backend/internal/handlers"
	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/models"
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
	patientService := services.NewPatientService(database.DB)
	documentService := services.NewDocumentService(database.DB)
	prescriptionService := services.NewPrescriptionService(database.DB)
	anamneseService := services.NewAnamneseService(database.DB)
	productService := services.NewProductService(database.DB)
	stockService := services.NewStockService(database.DB)
	orderService := services.NewOrderService(database.DB)
	financialService := services.NewFinancialService(database.DB)
	dashboardService := services.NewDashboardService(database.DB)

	authHandler := handlers.NewAuthHandler(authService)
	doctorHandler := handlers.NewDoctorHandler(doctorService)
	patientHandler := handlers.NewPatientHandler(patientService)
	documentHandler := handlers.NewDocumentHandler(documentService)
	prescriptionHandler := handlers.NewPrescriptionHandler(prescriptionService)
	anamneseHandler := handlers.NewAnamneseHandler(anamneseService)
	productHandler := handlers.NewProductHandler(productService)
	stockHandler := handlers.NewStockHandler(stockService)
	orderHandler := handlers.NewOrderHandler(orderService)
	financialHandler := handlers.NewFinancialHandler(financialService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	// ============================================================
	// PASSO 5: Configurar rotas
	// ============================================================
	log.Println("🛣️ Configurando rotas...")
	r := chi.NewRouter()

	// Middlewares globais
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.LoggerMiddleware) // nosso logger
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// ============================================================
	// 🆕 SERVER ARQUIVOS ESTÁTICOS (UPLOADS)
	// ============================================================
	// Isso permite que os arquivos sejam acessados via URL
	// Exemplo: http://localhost:8080/uploads/receitas/arquivo.pdf
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// ============================================================
	// ROTAS PÚBLICAS
	// ============================================================
	r.Get("/health", healthCheckHandler)
	r.Get("/", welcomeHandler)
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

		// --- Perfil do Usuário ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(jwtService))
			r.Get("/api/users/me", func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value(middleware.UserIDKey).(string)
				// Buscar usuário no banco
				var user models.User
				if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
					utils.SendError(w, http.StatusNotFound, "Usuário não encontrado")
					return
				}
				utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
					"id":         user.ID,
					"name":       user.Name,
					"email":      user.Email,
					"role":       user.Role,
					"is_active":  user.IsActive,
					"created_at": user.CreatedAt,
				})
			})
		})
		// --- Médicos (admin, secretaria, coordenacao) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao"))
			r.Post("/api/doctors", doctorHandler.Create)
			r.Get("/api/doctors", doctorHandler.List)
			r.Get("/api/doctors/top", doctorHandler.GetTopDoctors)
			r.Get("/api/doctors/{id}", doctorHandler.GetByID)
			r.Put("/api/doctors/{id}", doctorHandler.Update)
			r.Delete("/api/doctors/{id}", doctorHandler.Delete)
		})

		// --- Pacientes (admin, secretaria, coordenacao, acolhimento) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "acolhimento"))
			r.Post("/api/patients", patientHandler.Create)
			r.Get("/api/patients", patientHandler.List)
			r.Get("/api/patients/stats", patientHandler.GetStatistics)
			r.Get("/api/patients/{id}", patientHandler.GetByID)
			r.Put("/api/patients/{id}", patientHandler.Update)
			r.Patch("/api/patients/{id}/status", patientHandler.UpdateStatus)
			r.Delete("/api/patients/{id}", patientHandler.Delete)
		})

		// --- Documentos (admin, secretaria, coordenacao, acolhimento) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "acolhimento"))
			r.Post("/api/patients/{id}/documents", documentHandler.Upload)
			r.Get("/api/patients/{id}/documents", documentHandler.ListByPatient)
			r.Get("/api/documents/{id}", documentHandler.GetByID)
			r.Get("/api/documents/{id}/download", documentHandler.Download)
			r.Patch("/api/documents/{id}/status", documentHandler.UpdateStatus)
			r.Delete("/api/documents/{id}", documentHandler.Delete)
		})

		// --- Prescrições (admin, secretaria, coordenacao, acolhimento) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "acolhimento"))
			r.Post("/api/prescriptions", prescriptionHandler.Create)
			r.Get("/api/prescriptions", prescriptionHandler.List)
			r.Get("/api/prescriptions/expired", prescriptionHandler.GetExpired)
			r.Get("/api/prescriptions/validate/{id}", prescriptionHandler.Validate)
			r.Get("/api/prescriptions/{id}", prescriptionHandler.GetByID)
			r.Put("/api/prescriptions/{id}", prescriptionHandler.Update)
			r.Delete("/api/prescriptions/{id}", prescriptionHandler.Delete)
			r.Post("/api/prescriptions/update-status", prescriptionHandler.UpdateAllStatus)
		})

		// --- Anamnese (admin, secretaria, coordenacao, acolhimento) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "acolhimento"))
			r.Post("/api/patients/{id}/anamnesis", anamneseHandler.Create)
			r.Get("/api/patients/{id}/anamnesis", anamneseHandler.GetByPatient)
			r.Get("/api/anamnesis", anamneseHandler.List)
			r.Get("/api/anamnesis/{id}", anamneseHandler.GetByID)
			r.Put("/api/anamnesis/{id}", anamneseHandler.Update)
			r.Delete("/api/anamnesis/{id}", anamneseHandler.Delete)
		})

		// --- Produtos (admin, secretaria, coordenacao) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao"))
			r.Post("/api/products", productHandler.Create)
			r.Get("/api/products", productHandler.List)
			r.Get("/api/products/low-stock", productHandler.GetLowStock)
			r.Get("/api/products/stock-summary", productHandler.GetStockSummary)
			r.Get("/api/products/{id}", productHandler.GetByID)
			r.Put("/api/products/{id}", productHandler.Update)
			r.Delete("/api/products/{id}", productHandler.Delete)
		})

		// --- Estoque (admin, secretaria, coordenacao, farmacia) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "farmacia"))
			r.Post("/api/stock/lots", stockHandler.CreateLot)
			r.Get("/api/stock/lots", stockHandler.ListLots)
			r.Get("/api/stock/lots/{id}", stockHandler.GetLotByID)
			r.Post("/api/stock/adjust", stockHandler.AdjustStock)
			r.Get("/api/stock/movements", stockHandler.GetMovements)
			r.Get("/api/stock/expiring", stockHandler.GetExpiringLots)
			r.Get("/api/stock/low-stock", stockHandler.GetLowStock)
			r.Get("/api/stock/summary", stockHandler.GetStockSummary)
		})

		// --- Pedidos (admin, secretaria, coordenacao, farmacia) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao", "farmacia"))
			r.Post("/api/orders", orderHandler.Create)
			r.Get("/api/orders", orderHandler.List)
			r.Get("/api/orders/{id}", orderHandler.GetByID)
			r.Get("/api/orders/patient/{id}", orderHandler.GetByPatient)
			r.Patch("/api/orders/{id}/status", orderHandler.UpdateStatus)
			r.Patch("/api/orders/{id}/tracking", orderHandler.UpdateTracking)
			r.Post("/api/orders/{id}/label", orderHandler.GenerateLabel)
		})

		// --- Financeiro (admin, secretaria, coordenacao) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "secretaria", "coordenacao"))
			r.Post("/api/financial/subscriptions", financialHandler.CreateSubscription)
			r.Get("/api/financial/subscriptions", financialHandler.ListSubscriptions)
			r.Get("/api/financial/subscriptions/{id}", financialHandler.GetSubscriptionByID)
			r.Post("/api/financial/payments", financialHandler.CreatePayment)
			r.Get("/api/financial/payments", financialHandler.ListPayments)
			r.Get("/api/financial/payments/{id}", financialHandler.GetPaymentByID)
			r.Patch("/api/financial/payments/{id}/status", financialHandler.UpdatePaymentStatus)
			r.Get("/api/financial/patient/{id}", financialHandler.GetPatientFinancialStatus)
			r.Get("/api/financial/overdue", financialHandler.GetOverdueSubscriptions)
		})

		// --- Dashboard (admin, coordenacao) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("admin", "coordenacao"))
			r.Get("/api/dashboard/overview", dashboardHandler.GetOverview)
			r.Get("/api/dashboard/patients", dashboardHandler.GetPatientReport)
			r.Get("/api/dashboard/expired-prescriptions", dashboardHandler.GetExpiredPrescriptions)
			r.Get("/api/dashboard/top-doctors", dashboardHandler.GetTopDoctors)
			r.Get("/api/dashboard/low-stock", dashboardHandler.GetLowStock)
		})

		// --- Admin Only ---
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
	log.Println("")
	log.Println("📋 Documentação da API:")
	log.Println("   🔐 Auth:      /api/auth/register, /api/auth/login")
	log.Println("   👨‍⚕️ Doctors:   /api/doctors")
	log.Println("   👤 Patients:  /api/patients")
	log.Println("   📄 Documents: /api/patients/{id}/documents")
	log.Println("   💊 Prescriptions: /api/prescriptions")
	log.Println("   📝 Anamnese:  /api/patients/{id}/anamnesis")
	log.Println("   📦 Products:  /api/products")
	log.Println("   📊 Stock:     /api/stock")
	log.Println("   🛒 Orders:    /api/orders")
	log.Println("   💰 Financial: /api/financial")
	log.Println("   📈 Dashboard: /api/dashboard")
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
		"etapa":    "14 - Middleware e Permissões",
	})
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "🌿 Bem-vindo à API CannaCare!",
		"version": "1.0.0",
		"etapa":   "14 - Middleware e Permissões",
		"documentation": map[string]string{
			"POST /api/auth/register":      "Registrar usuário",
			"POST /api/auth/login":         "Login",
			"GET  /api/doctors":            "Listar médicos",
			"GET  /api/patients":           "Listar pacientes",
			"GET  /api/products":           "Listar produtos",
			"GET  /api/stock/lots":         "Listar lotes",
			"GET  /api/orders":             "Listar pedidos",
			"GET  /api/financial/payments": "Listar pagamentos",
			"GET  /api/dashboard/overview": "Dashboard",
		},
	})
}
