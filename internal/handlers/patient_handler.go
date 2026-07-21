// ================================================================
// PACOTE HANDLERS - PATIENT HANDLER
// ================================================================
// Camada HTTP que lida com as requisições de pacientes.
//
// ENDPOINTS:
//   POST   /api/patients           - Criar paciente
//   GET    /api/patients           - Listar pacientes (com filtros)
//   GET    /api/patients/{id}      - Buscar paciente por ID
//   PUT    /api/patients/{id}      - Atualizar paciente
//   DELETE /api/patients/{id}      - Remover paciente (soft delete)
//   PATCH  /api/patients/{id}/status - Mudar status do paciente
//   GET    /api/patients/stats     - Estatísticas de pacientes
// ================================================================

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ================================================================
// STRUCT PATIENTHANDLER
// ================================================================
type PatientHandler struct {
	patientService *services.PatientService
	validator      *validator.Validate
}

// ================================================================
// FUNÇÃO NEWPATIENTHANDLER()
// ================================================================
func NewPatientHandler(patientService *services.PatientService) *PatientHandler {
	return &PatientHandler{
		patientService: patientService,
		validator:      validator.New(),
	}
}

// ================================================================
// HANDLER: CREATE
// ================================================================
// Endpoint: POST /api/patients
func (h *PatientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req services.CreatePatientRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	patient, err := h.patientService.Create(req)
	if err != nil {
		log.Printf("❌ Erro ao criar paciente: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, patient)
}

// ================================================================
// HANDLER: GET BY ID
// ================================================================
// Endpoint: GET /api/patients/{id}
func (h *PatientHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	patient, err := h.patientService.GetByID(id)
	if err != nil {
		if err.Error() == "paciente não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar paciente: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar paciente")
		return
	}

	utils.SendSuccess(w, http.StatusOK, patient)
}

// ================================================================
// HANDLER: LIST
// ================================================================
// Endpoint: GET /api/patients
func (h *PatientHandler) List(w http.ResponseWriter, r *http.Request) {
	req := services.ListPatientRequest{
		Name:   r.URL.Query().Get("name"),
		CPF:    r.URL.Query().Get("cpf"),
		Email:  r.URL.Query().Get("email"),
		Status: r.URL.Query().Get("status"),
	}

	if isSocialStr := r.URL.Query().Get("is_social"); isSocialStr != "" {
		isSocial, err := strconv.ParseBool(isSocialStr)
		if err == nil {
			req.IsSocial = &isSocial
		}
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil {
			req.Page = page
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil {
			req.Limit = limit
		}
	}

	patients, total, err := h.patientService.List(req)
	if err != nil {
		log.Printf("❌ Erro ao listar pacientes: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar pacientes")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": patients,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// ================================================================
// HANDLER: UPDATE
// ================================================================
// Endpoint: PUT /api/patients/{id}
func (h *PatientHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req services.UpdatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	patient, err := h.patientService.Update(id, req)
	if err != nil {
		if err.Error() == "paciente não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar paciente: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, patient)
}

// ================================================================
// HANDLER: DELETE
// ================================================================
// Endpoint: DELETE /api/patients/{id}
func (h *PatientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.patientService.Delete(id); err != nil {
		if err.Error() == "paciente não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao deletar paciente: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao deletar paciente")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Paciente removido com sucesso",
	})
}

// ================================================================
// HANDLER: UPDATE STATUS
// ================================================================
// Endpoint: PATCH /api/patients/{id}/status
func (h *PatientHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req services.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "status inválido: "+err.Error())
		return
	}

	// Obter ID do usuário que está fazendo a alteração
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID do usuário")
		return
	}

	patient, err := h.patientService.UpdateStatus(id, req, userID)
	if err != nil {
		if err.Error() == "paciente não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar status: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, patient)
}

// ================================================================
// HANDLER: GET STATISTICS
// ================================================================
// Endpoint: GET /api/patients/stats
func (h *PatientHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.patientService.GetStatistics()
	if err != nil {
		log.Printf("❌ Erro ao buscar estatísticas: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar estatísticas")
		return
	}

	utils.SendSuccess(w, http.StatusOK, stats)
}
