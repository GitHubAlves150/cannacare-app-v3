// ================================================================
// PACOTE HANDLERS - DASHBOARD HANDLER
// ================================================================
// Camada HTTP que lida com as requisições de dashboard e relatórios.
//
// ENDPOINTS:
//   GET /api/dashboard/overview           - Visão geral do sistema
//   GET /api/dashboard/patients           - Relatório de pacientes
//   GET /api/dashboard/expired-prescriptions - Receitas vencidas
//   GET /api/dashboard/top-doctors        - Médicos que mais prescrevem
//   GET /api/dashboard/low-stock          - Produtos com estoque baixo
// ================================================================

package handlers

import (
	"log"
	"net/http"

	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"
)

// ================================================================
// STRUCT DASHBOARDHANDLER
// ================================================================
type DashboardHandler struct {
	dashboardService *services.DashboardService
}

// ================================================================
// FUNÇÃO NEWDASHBOARDHANDLER()
// ================================================================
func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// ================================================================
// HANDLER: GET OVERVIEW
// ================================================================
// Endpoint: GET /api/dashboard/overview
func (h *DashboardHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.dashboardService.GetOverview()
	if err != nil {
		log.Printf("❌ Erro ao buscar overview: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar dados do dashboard")
		return
	}

	utils.SendSuccess(w, http.StatusOK, overview)
}

// ================================================================
// HANDLER: GET PATIENT REPORT
// ================================================================
// Endpoint: GET /api/dashboard/patients
func (h *DashboardHandler) GetPatientReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.dashboardService.GetPatientReport()
	if err != nil {
		log.Printf("❌ Erro ao buscar relatório de pacientes: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar relatório de pacientes")
		return
	}

	utils.SendSuccess(w, http.StatusOK, report)
}

// ================================================================
// HANDLER: GET EXPIRED PRESCRIPTIONS
// ================================================================
// Endpoint: GET /api/dashboard/expired-prescriptions
func (h *DashboardHandler) GetExpiredPrescriptions(w http.ResponseWriter, r *http.Request) {
	report, err := h.dashboardService.GetExpiredPrescriptionsReport()
	if err != nil {
		log.Printf("❌ Erro ao buscar receitas vencidas: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar receitas vencidas")
		return
	}

	utils.SendSuccess(w, http.StatusOK, report)
}

// ================================================================
// HANDLER: GET TOP DOCTORS
// ================================================================
// Endpoint: GET /api/dashboard/top-doctors
func (h *DashboardHandler) GetTopDoctors(w http.ResponseWriter, r *http.Request) {
	report, err := h.dashboardService.GetTopDoctorsReport()
	if err != nil {
		log.Printf("❌ Erro ao buscar top médicos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar top médicos")
		return
	}

	utils.SendSuccess(w, http.StatusOK, report)
}

// ================================================================
// HANDLER: GET LOW STOCK
// ================================================================
// Endpoint: GET /api/dashboard/low-stock
func (h *DashboardHandler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	report, err := h.dashboardService.GetLowStockReport()
	if err != nil {
		log.Printf("❌ Erro ao buscar produtos com estoque baixo: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar produtos com estoque baixo")
		return
	}

	utils.SendSuccess(w, http.StatusOK, report)
}
