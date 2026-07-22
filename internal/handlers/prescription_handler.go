// ================================================================
// PACOTE HANDLERS - PRESCRIPTION HANDLER
// ================================================================
// Camada HTTP que lida com as requisições de prescrições.
//
// ENDPOINTS:
//   POST   /api/prescriptions           - Criar prescrição
//   GET    /api/prescriptions           - Listar prescrições
//   GET    /api/prescriptions/{id}      - Buscar prescrição por ID
//   PUT    /api/prescriptions/{id}      - Atualizar prescrição
//   DELETE /api/prescriptions/{id}      - Remover prescrição
//   GET    /api/prescriptions/validate/{id} - Validar prescrição
//   GET    /api/prescriptions/expired   - Prescrições vencidas
//   POST   /api/prescriptions/update-status - Atualizar status em lote
// ================================================================

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ================================================================
// STRUCT PRESCRIPTIONHANDLER
// ================================================================
type PrescriptionHandler struct {
	prescriptionService *services.PrescriptionService
	validator           *validator.Validate
}

// ================================================================
// FUNÇÃO NEWPRESCRIPTIONHANDLER()
// ================================================================
func NewPrescriptionHandler(prescriptionService *services.PrescriptionService) *PrescriptionHandler {
	return &PrescriptionHandler{
		prescriptionService: prescriptionService,
		validator:           validator.New(),
	}
}

// ================================================================
// HANDLER: CREATE
// ================================================================
// Endpoint: POST /api/prescriptions
func (h *PrescriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req services.CreatePrescriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	prescription, err := h.prescriptionService.Create(req)
	if err != nil {
		log.Printf("❌ Erro ao criar prescrição: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, prescription)
}

// ================================================================
// HANDLER: GET BY ID
// ================================================================
// Endpoint: GET /api/prescriptions/{id}
func (h *PrescriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	prescription, err := h.prescriptionService.GetByID(id)
	if err != nil {
		if err.Error() == "prescrição não encontrada" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar prescrição: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar prescrição")
		return
	}

	utils.SendSuccess(w, http.StatusOK, prescription)
}

// ================================================================
// HANDLER: LIST
// ================================================================
// Endpoint: GET /api/prescriptions
func (h *PrescriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	req := services.ListPrescriptionRequest{
		PatientID: r.URL.Query().Get("patient_id"),
		DoctorID:  r.URL.Query().Get("doctor_id"),
		Status:    r.URL.Query().Get("status"),
	}

	if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			req.IsActive = &isActive
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

	prescriptions, total, err := h.prescriptionService.List(req)
	if err != nil {
		log.Printf("❌ Erro ao listar prescrições: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar prescrições")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": prescriptions,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// ================================================================
// HANDLER: UPDATE
// ================================================================
// Endpoint: PUT /api/prescriptions/{id}
func (h *PrescriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req services.UpdatePrescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	prescription, err := h.prescriptionService.Update(id, req)
	if err != nil {
		if err.Error() == "prescrição não encontrada" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar prescrição: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, prescription)
}

// ================================================================
// HANDLER: DELETE
// ================================================================
// Endpoint: DELETE /api/prescriptions/{id}
func (h *PrescriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.prescriptionService.Delete(id); err != nil {
		if err.Error() == "prescrição não encontrada" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao deletar prescrição: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao deletar prescrição")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Prescrição removida com sucesso",
	})
}

// ================================================================
// HANDLER: VALIDATE
// ================================================================
// Endpoint: GET /api/prescriptions/validate/{id}
func (h *PrescriptionHandler) Validate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	result, err := h.prescriptionService.Validate(id)
	if err != nil {
		log.Printf("❌ Erro ao validar prescrição: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao validar prescrição")
		return
	}

	utils.SendSuccess(w, http.StatusOK, result)
}

// ================================================================
// HANDLER: GET EXPIRED
// ================================================================
// Endpoint: GET /api/prescriptions/expired
func (h *PrescriptionHandler) GetExpired(w http.ResponseWriter, r *http.Request) {
	prescriptions, err := h.prescriptionService.GetExpired()
	if err != nil {
		log.Printf("❌ Erro ao buscar prescrições vencidas: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar prescrições vencidas")
		return
	}

	utils.SendSuccess(w, http.StatusOK, prescriptions)
}

// ================================================================
// HANDLER: UPDATE STATUS (JOB)
// ================================================================
// Endpoint: POST /api/prescriptions/update-status
func (h *PrescriptionHandler) UpdateAllStatus(w http.ResponseWriter, r *http.Request) {
	if err := h.prescriptionService.UpdateAllStatus(); err != nil {
		log.Printf("❌ Erro ao atualizar status das prescrições: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao atualizar status das prescrições")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Status das prescrições atualizados com sucesso",
	})
}
