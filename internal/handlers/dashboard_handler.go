// ================================================================
// PACOTE HANDLERS - DASHBOARD HANDLER
// ================================================================
package handlers

import (
	"log"
	"net/http"

	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/google/uuid"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) extractAssociationID(r *http.Request) (uuid.UUID, error) {
	associationIDStr, _ := r.Context().Value(middleware.AssociationIDKey).(string)
	if associationIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(associationIDStr)
}

// GET /api/dashboard/overview
func (h *DashboardHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	overview, err := h.dashboardService.GetOverview(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar overview: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar dados do dashboard")
		return
	}

	utils.SendSuccess(w, http.StatusOK, overview)
}

// GET /api/dashboard/patients
func (h *DashboardHandler) GetPatientReport(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	report, err := h.dashboardService.GetPatientReport(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar relatório de pacientes: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar relatório de pacientes")
		return
	}

	utils.SendSuccess(w, http.StatusOK, report)
}

// GET /api/dashboard/expired-prescriptions
func (h *DashboardHandler) GetExpiredPrescriptions(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	report, err := h.dashboardService.GetExpiredPrescriptionsReport(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar receitas vencidas: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar receitas vencidas")
		return
	}

	utils.SendSuccess(w, http.StatusOK, report)
}

// GET /api/dashboard/top-doctors
func (h *DashboardHandler) GetTopDoctors(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	report, err := h.dashboardService.GetTopDoctorsReport(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar top médicos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar top médicos")
		return
	}

	utils.SendSuccess(w, http.StatusOK, report)
}

// GET /api/dashboard/low-stock
func (h *DashboardHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	report, err := h.dashboardService.GetLowStockReport(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar produtos com estoque baixo: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar produtos com estoque baixo")
		return
	}

	utils.SendSuccess(w, http.StatusOK, report)
}