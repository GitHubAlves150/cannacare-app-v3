// ================================================================
// PACOTE SERVICES - PLAN LIFECYCLE SERVICE
// ================================================================
// Roda uma vez por dia (chamado a partir de uma goroutine no main.go).
// Duas responsabilidades:
//
//   1. CheckExpiringPlans  - avisa (in-app + email) associações premium
//      cujo plano vai vencer em 15, 7 ou 1 dia
//   2. DowngradeExpiredPlans - rebaixa pra 'basic' quem já passou da
//      data de expiração, e avisa por email
//
// Idempotência: cada aviso in-app carrega o número de dias restantes
// no título (ex: "expira em 7 dias"), então não repete o mesmo aviso
// duas vezes — se já existe uma notificação com esse título pra essa
// associação, pula.
// ================================================================

package services

import (
	"fmt"
	"log"
	"time"

	"cannacare-backend/internal/models"

	"gorm.io/gorm"
)

type PlanLifecycleService struct {
	db           *gorm.DB
	emailService *EmailService
}

func NewPlanLifecycleService(db *gorm.DB, emailService *EmailService) *PlanLifecycleService {
	return &PlanLifecycleService{db: db, emailService: emailService}
}

// Limiares de aviso, em dias antes de expirar.
var avisoThresholds = []int{15, 7, 1}

// ================================================================
// RUNDAILYCHECKS - ponto de entrada único, chamado pelo scheduler
// ================================================================
func (s *PlanLifecycleService) RunDailyChecks() {
	log.Println("🔔 Rodando verificação diária de planos...")

	if err := s.CheckExpiringPlans(); err != nil {
		log.Printf("❌ erro ao verificar planos expirando: %v", err)
	}
	if err := s.DowngradeExpiredPlans(); err != nil {
		log.Printf("❌ erro ao rebaixar planos expirados: %v", err)
	}

	log.Println("✅ Verificação diária de planos concluída")
}

// ================================================================
// CHECKEXPIRINGPLANS
// ================================================================
func (s *PlanLifecycleService) CheckExpiringPlans() error {
	var associations []models.Association
	maxWindow := time.Now().AddDate(0, 0, avisoThresholds[0]) // maior limiar (15 dias)

	if err := s.db.
		Where("plan = ? AND status = ? AND plan_expires_at IS NOT NULL", "premium", "active").
		Where("plan_expires_at BETWEEN ? AND ?", time.Now(), maxWindow).
		Find(&associations).Error; err != nil {
		return err
	}

	for _, association := range associations {
		daysLeft := int(time.Until(*association.PlanExpiresAt).Hours() / 24)

		for _, threshold := range avisoThresholds {
			if daysLeft == threshold {
				s.notifyExpiringSoon(association, threshold)
			}
		}
	}

	return nil
}

func (s *PlanLifecycleService) notifyExpiringSoon(association models.Association, daysLeft int) {
	title := fmt.Sprintf("Plano expira em %d dias", daysLeft)
	message := fmt.Sprintf("O plano premium da associação expira em %s. Renove para não perder os recursos do plano.",
		association.PlanExpiresAt.Format("02/01/2006"))

	// --- Notificação in-app pra cada admin da associação ---
	var admins []models.User
	if err := s.db.Where("association_id = ? AND role = ? AND is_active = ?", association.ID, "admin", true).
		Find(&admins).Error; err != nil {
		log.Printf("⚠️ erro ao buscar admins da associação %s: %v", association.ID, err)
		return
	}

	for _, admin := range admins {
		// Idempotência: já existe um aviso com esse título pra esse admin?
		var count int64
		s.db.Model(&models.Notification{}).
			Where("association_id = ? AND user_id = ? AND title = ?", association.ID, admin.ID, title).
			Count(&count)
		if count > 0 {
			continue
		}

		notification := &models.Notification{
			AssociationID: association.ID,
			UserID:        admin.ID,
			Type:          "plan_expiring",
			Title:         title,
			Message:       message,
			ActionURL:     "/dashboard/admin/plano",
		}
		if err := s.db.Create(notification).Error; err != nil {
			log.Printf("⚠️ erro ao criar notificação para %s: %v", admin.Email, err)
			continue
		}

		// --- Email (só uma vez, mesmo helper de idempotência acima) ---
		if err := s.emailService.SendPlanExpiringEmail(admin.Email, admin.Name, association.Name, daysLeft, *association.PlanExpiresAt); err != nil {
			log.Printf("⚠️ erro ao enviar email de expiração para %s: %v", admin.Email, err)
		}
	}
}

// ================================================================
// DOWNGRADEEXPIREDPLANS
// ================================================================
func (s *PlanLifecycleService) DowngradeExpiredPlans() error {
	var associations []models.Association
	if err := s.db.
		Where("plan = ? AND status = ? AND plan_expires_at IS NOT NULL AND plan_expires_at < ?", "premium", "active", time.Now()).
		Find(&associations).Error; err != nil {
		return err
	}

	for _, association := range associations {
		association.Plan = "basic"
		association.PatientLimit = 50
		association.UserLimit = 3
		// Mantém Status = "active" — a associação continua existindo e
		// usável, só que agora com os limites do plano básico.

		if err := s.db.Save(&association).Error; err != nil {
			log.Printf("⚠️ erro ao rebaixar associação %s: %v", association.ID, err)
			continue
		}

		log.Printf("⬇️ Associação %s (%s) rebaixada para o plano básico (plano expirou)", association.Name, association.ID)

		var admins []models.User
		s.db.Where("association_id = ? AND role = ? AND is_active = ?", association.ID, "admin", true).Find(&admins)

		for _, admin := range admins {
			notification := &models.Notification{
				AssociationID: association.ID,
				UserID:        admin.ID,
				Type:          "plan_expired",
				Title:         "Plano expirou",
				Message:       "O plano premium expirou e a conta voltou para o plano básico. Renove para recuperar os recursos do plano premium.",
				ActionURL:     "/dashboard/admin/plano",
			}
			if err := s.db.Create(notification).Error; err != nil {
				log.Printf("⚠️ erro ao criar notificação de expiração para %s: %v", admin.Email, err)
			}

			if err := s.emailService.SendPlanExpiredEmail(admin.Email, admin.Name, association.Name); err != nil {
				log.Printf("⚠️ erro ao enviar email de plano expirado para %s: %v", admin.Email, err)
			}
		}
	}

	return nil
}