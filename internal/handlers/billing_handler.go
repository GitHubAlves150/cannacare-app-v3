// ================================================================
// PACOTE HANDLERS - BILLING HANDLER (AUTENTICADO)
// ================================================================
// Endpoints que precisam de login (diferente de public_handler.go).
//
//   POST /api/billing/renew        - gera checkout de renovação
//   GET  /api/notifications        - lista notificações do usuário logado
//   PATCH /api/notifications/{id}/read - marca como lida
// ================================================================

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/models"
	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillingHandler struct {
	onboardingService *services.OnboardingService
	db                *gorm.DB
}

func NewBillingHandler(onboardingService *services.OnboardingService, db *gorm.DB) *BillingHandler {
	return &BillingHandler{onboardingService: onboardingService, db: db}
}

func (h *BillingHandler) extractAssociationID(r *http.Request) (uuid.UUID, error) {
	associationIDStr, _ := r.Context().Value(middleware.AssociationIDKey).(string)
	if associationIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(associationIDStr)
}

func (h *BillingHandler) extractUserID(r *http.Request) (uuid.UUID, error) {
	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	if userIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(userIDStr)
}

// ================================================================
// GET /api/billing/plan
// ================================================================
// Dados do plano atual da associação — alimenta a tela "Meu Plano".
func (h *BillingHandler) GetPlanInfo(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	var association models.Association
	if err := h.db.Where("id = ?", associationID).First(&association).Error; err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar dados do plano")
		return
	}

	var patientCount int64
	h.db.Model(&models.Patient{}).Where("association_id = ?", associationID).Count(&patientCount)

	var userCount int64
	h.db.Model(&models.User{}).Where("association_id = ? AND is_active = ?", associationID, true).Count(&userCount)

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"plan":              association.Plan,
		"status":            association.Status,
		"plan_activated_at": association.PlanActivatedAt,
		"plan_expires_at":   association.PlanExpiresAt,
		"patient_limit":     association.PatientLimit,
		"user_limit":        association.UserLimit,
		"patient_count":     patientCount,
		"user_count":        userCount,
	})
}

// ================================================================
// POST /api/billing/renew
// ================================================================
func (h *BillingHandler) RenewPremium(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	checkoutURL, err := h.onboardingService.CreateRenewalCheckout(associationID)
	if err != nil {
		log.Printf("❌ erro ao gerar checkout de renovação: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"checkout_url": checkoutURL,
	})
}

// ================================================================
// GET /api/notifications
// ================================================================
func (h *BillingHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil || userID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "usuário não identificado")
		return
	}

	var notifications []models.Notification
	if err := h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50).
		Find(&notifications).Error; err != nil {
		log.Printf("❌ erro ao listar notificações: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar notificações")
		return
	}

	var unreadCount int64
	h.db.Model(&models.Notification{}).Where("user_id = ? AND read_at IS NULL", userID).Count(&unreadCount)

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items":        notifications,
		"unread_count": unreadCount,
	})
}

// ================================================================
// PATCH /api/notifications/{id}/read
// ================================================================
func (h *BillingHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil || userID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "usuário não identificado")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var notification models.Notification
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&notification).Error; err != nil {
		utils.SendError(w, http.StatusNotFound, "notificação não encontrada")
		return
	}

	if notification.ReadAt == nil {
		var body struct{}
		_ = json.NewDecoder(r.Body).Decode(&body) // corpo opcional, não usamos nada dele

		h.db.Model(&notification).Update("read_at", gorm.Expr("now()"))
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{"message": "ok"})
}