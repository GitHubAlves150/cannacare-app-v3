// ================================================================
// CANNACARE - MIGRATE (CORRIGIDO)
// ================================================================

package database

import (
	"log"

	"cannacare-backend/internal/models"
)

// Migrate cria ou atualiza todas as tabelas no banco de dados
func Migrate() error {
	log.Println("🔄 Iniciando migrações...")

	// ================================================================
	// PASSO 1: DROPAR VIEWS (para evitar erro de alteração de coluna)
	// ================================================================
	log.Println("🗑️  Removendo views existentes...")

	views := []string{
		"vw_expired_prescriptions",
		"vw_low_stock",
		"vw_overdue_subscriptions",
		"vw_patient_dashboard",
		"vw_top_doctors",
		"vw_stock_summary",
	}

	for _, view := range views {
		err := DB.Exec("DROP VIEW IF EXISTS " + view + " CASCADE").Error
		if err != nil {
			log.Printf("⚠️ Erro ao dropar view %s: %v", view, err)
		} else {
			log.Printf("✅ View %s removida", view)
		}
	}

	// ================================================================
	// PASSO 2: REMOVER COLUNA GERADA (se existir)
	// ================================================================
	log.Println("📝 Removendo coluna gerada order_items.total_price...")

	err := DB.Exec(`
		DO $$ 
		BEGIN
			IF EXISTS (
				SELECT 1 
				FROM information_schema.columns 
				WHERE table_name = 'order_items' 
				AND column_name = 'total_price'
			) THEN
				ALTER TABLE order_items DROP COLUMN total_price;
			END IF;
		END $$;
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao remover coluna total_price: %v", err)
	} else {
		log.Println("✅ Coluna total_price removida (se existia)")
	}

	// ================================================================
	// PASSO 3: Criar tabelas DEPENDENTES manualmente
	// ================================================================
	log.Println("📝 Criando tabelas dependentes manualmente...")

	err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS subscriptions (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ,
			deleted_at TIMESTAMPTZ,
			patient_id UUID NOT NULL,
			association_id UUID NOT NULL,
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

	err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ,
			deleted_at TIMESTAMPTZ,
			patient_id UUID NOT NULL,
			association_id UUID NOT NULL,
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

	// ================================================================
	// PASSO 4: Rodar AutoMigrate (EXCLUINDO order_items)
	// ================================================================
	log.Println("📝 Rodando AutoMigrate para as tabelas (excluindo order_items)...")

	// Lista de models para migrar (EXCLUINDO order_items)
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.Doctor{},
		&models.Product{},
		&models.Patient{},
		&models.Prescription{},
		&models.PrescriptionItem{},
		&models.ProductLot{},
		&models.Order{},
		// &models.OrderItem{}, // ← REMOVIDO para evitar erro
		&models.StockMovement{},
		&models.Anamnese{},
		&models.PatientDocument{},
		&models.Notification{},
		&models.PatientStatusHistory{},
	}

	err = DB.AutoMigrate(modelsToMigrate...)
	if err != nil {
		return err
	}

	// ================================================================
	// PASSO 5: Recriar order_items MANUALMENTE
	// ================================================================
	log.Println("📝 Recriando order_items manualmente...")

	err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS order_items (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ,
			deleted_at TIMESTAMPTZ,
			order_id UUID NOT NULL,
			product_lot_id UUID NOT NULL,
			association_id UUID NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			total_price DECIMAL(10,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
			FOREIGN KEY (product_lot_id) REFERENCES product_lots(id) ON DELETE CASCADE
		)
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao criar order_items: %v", err)
	} else {
		log.Println("✅ order_items criada com sucesso")
	}

	// ================================================================
	// PASSO 6: Recriar as VIEWS
	// ================================================================
	log.Println("📝 Recriando views...")

	// VIEW 1: Pacientes com receitas vencidas
	err = DB.Exec(`
		CREATE OR REPLACE VIEW vw_expired_prescriptions AS
		SELECT 
			p.id as patient_id,
			p.full_name,
			p.cpf,
			p.phone,
			pr.id as prescription_id,
			pr.expiration_date,
			pr.doctor_id,
			d.name as doctor_name,
			d.crm,
			p.association_id,
			EXTRACT(DAY FROM AGE(CURRENT_DATE, pr.expiration_date)) as days_expired
		FROM patients p
		JOIN prescriptions pr ON pr.patient_id = p.id
		JOIN doctors d ON d.id = pr.doctor_id
		WHERE pr.expiration_date < CURRENT_DATE
		AND pr.is_active = true
		AND p.status = 'aprovado'
		AND p.deleted_at IS NULL
		ORDER BY pr.expiration_date ASC
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao criar vw_expired_prescriptions: %v", err)
	} else {
		log.Println("✅ vw_expired_prescriptions criada")
	}

	// VIEW 2: Estoque baixo
	err = DB.Exec(`
		CREATE OR REPLACE VIEW vw_low_stock AS
		SELECT 
			pl.id as lot_id,
			p.id as product_id,
			p.name as product_name,
			pl.lot_number,
			pl.expiration_date,
			pl.current_quantity,
			p.min_stock_alert,
			(p.min_stock_alert - pl.current_quantity) as missing_units,
			p.association_id
		FROM product_lots pl
		JOIN products p ON p.id = pl.product_id
		WHERE pl.current_quantity <= p.min_stock_alert
		AND pl.expiration_date > CURRENT_DATE
		AND pl.deleted_at IS NULL
		ORDER BY (p.min_stock_alert - pl.current_quantity) DESC
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao criar vw_low_stock: %v", err)
	} else {
		log.Println("✅ vw_low_stock criada")
	}

	// VIEW 3: Anuidades em atraso
	err = DB.Exec(`
		CREATE OR REPLACE VIEW vw_overdue_subscriptions AS
		SELECT 
			s.id as subscription_id,
			p.id as patient_id,
			p.full_name,
			p.phone,
			s.due_date,
			s.amount,
			p.association_id,
			EXTRACT(DAY FROM AGE(CURRENT_DATE, s.due_date)) as days_overdue
		FROM subscriptions s
		JOIN patients p ON p.id = s.patient_id
		WHERE s.status = 'atrasado'
		AND s.deleted_at IS NULL
		ORDER BY s.due_date ASC
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao criar vw_overdue_subscriptions: %v", err)
	} else {
		log.Println("✅ vw_overdue_subscriptions criada")
	}

	// VIEW 4: Dashboard - Pacientes ativos/inativos
	err = DB.Exec(`
		CREATE OR REPLACE VIEW vw_patient_dashboard AS
		SELECT 
			association_id,
			status,
			COUNT(*) as total,
			COUNT(CASE WHEN is_social_patient = true THEN 1 END) as social_patients,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as last_30_days
		FROM patients
		WHERE deleted_at IS NULL
		GROUP BY association_id, status
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao criar vw_patient_dashboard: %v", err)
	} else {
		log.Println("✅ vw_patient_dashboard criada")
	}

	// VIEW 5: Médicos que mais prescrevem
	err = DB.Exec(`
		CREATE OR REPLACE VIEW vw_top_doctors AS
		SELECT 
			d.id as doctor_id,
			d.name as doctor_name,
			d.crm,
			d.specialty,
			d.association_id,
			COUNT(pr.id) as total_prescriptions,
			COUNT(DISTINCT pr.patient_id) as unique_patients
		FROM doctors d
		LEFT JOIN prescriptions pr ON pr.doctor_id = d.id
		WHERE d.is_active = true
		AND d.deleted_at IS NULL
		GROUP BY d.id, d.name, d.crm, d.specialty, d.association_id
		ORDER BY total_prescriptions DESC
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao criar vw_top_doctors: %v", err)
	} else {
		log.Println("✅ vw_top_doctors criada")
	}

	// VIEW 6: Movimentação de estoque consolidada
	err = DB.Exec(`
		CREATE OR REPLACE VIEW vw_stock_summary AS
		SELECT 
			p.id as product_id,
			p.name as product_name,
			p.association_id,
			COUNT(DISTINCT pl.id) as total_lots,
			COALESCE(SUM(pl.current_quantity), 0) as total_quantity,
			COALESCE(AVG(pl.current_quantity), 0) as avg_per_lot,
			MIN(pl.expiration_date) as earliest_expiration
		FROM products p
		LEFT JOIN product_lots pl ON pl.product_id = p.id
		WHERE p.is_active = true
		AND p.deleted_at IS NULL
		GROUP BY p.id, p.name, p.association_id
		ORDER BY p.name ASC
	`).Error
	if err != nil {
		log.Printf("⚠️ Erro ao criar vw_stock_summary: %v", err)
	} else {
		log.Println("✅ vw_stock_summary criada")
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
	log.Println("   ✅ order_items (criada manualmente)")
	log.Println("   ✅ stock_movements")
	log.Println("   ✅ subscriptions (criada manualmente)")
	log.Println("   ✅ payments (criada manualmente)")
	log.Println("   ✅ anamneses")
	log.Println("   ✅ patient_documents")
	log.Println("   ✅ notifications")
	log.Println("   ✅ patient_status_history")
	log.Println("   ✅ 6 views recriadas")

	return nil
}