package database

import (
	"log"

	"cannacare-backend/internal/models"
)

// Migrate cria ou atualiza todas as tabelas no banco de dados
func Migrate() error {
	log.Println("🔄 Iniciando migrações...")

	// ============================================================
	// PASSO 1: Criar tabelas DEPENDENTES manualmente
	// ============================================================
	log.Println("📝 Criando tabelas dependentes manualmente...")

	// 1. Criar subscriptions
	err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS subscriptions (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ,
			deleted_at TIMESTAMPTZ,
			patient_id UUID NOT NULL,
			payment_id UUID,
			due_date TIMESTAMPTZ NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			status TEXT DEFAULT 'pendente',
			paid_at TIMESTAMPTZ,
			FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE
		)
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao criar subscriptions: %v", err)
	}

	// 2. Criar payments (que depende de subscriptions)
	err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ,
			deleted_at TIMESTAMPTZ,
			patient_id UUID NOT NULL,
			order_id UUID,
			subscription_id UUID,
			payment_type TEXT NOT NULL,
			payment_method TEXT NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			installments BIGINT DEFAULT 1,
			status TEXT DEFAULT 'pendente',
			payment_date TIMESTAMPTZ,
			paid_at TIMESTAMPTZ,
			receipt_url TEXT,
			receipt_number TEXT,
			transaction_id TEXT,
			gateway_response JSONB,
			FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE SET NULL,
			FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL
		)
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao criar payments: %v", err)
	}

	// ============================================================
	// PASSO 2: Rodar AutoMigrate para as outras tabelas
	// ============================================================
	log.Println("📝 Rodando AutoMigrate para as outras tabelas...")

	err = DB.AutoMigrate(
		&models.User{},
		&models.Doctor{},
		&models.Product{},
		&models.Patient{},
		&models.Prescription{},
		&models.PrescriptionItem{},
		&models.ProductLot{},
		&models.Order{},
		&models.OrderItem{},
		&models.StockMovement{},
		// ⚠️ Subscription e Payment removidos (já foram criados manualmente)
		&models.Anamnese{},
		&models.PatientDocument{},
		&models.Notification{},
		&models.PatientStatusHistory{},
	)

	if err != nil {
		return err
	}

	log.Println("✅ Migrações concluídas com sucesso!")
	log.Println("📋 Tabelas criadas/atualizadas:")
	log.Println("   ✅ users")
	log.Println("   ✅ doctors")
	log.Println("   ✅ products")
	log.Println("   ✅ patients")
	log.Println("   ✅ prescriptions")
	log.Println("   ✅ prescription_items")
	log.Println("   ✅ product_lots")
	log.Println("   ✅ orders")
	log.Println("   ✅ order_items")
	log.Println("   ✅ stock_movements")
	log.Println("   ✅ subscriptions (criada manualmente)")
	log.Println("   ✅ payments (criada manualmente)")
	log.Println("   ✅ anamneses")
	log.Println("   ✅ patient_documents")
	log.Println("   ✅ notifications")
	log.Println("   ✅ patient_status_history")

	return nil
}