// ================================================================
// PACOTE HANDLERS - ANAMNESE HANDLER
// ================================================================
// Camada HTTP que lida com as requisições de anamnese.
//
// ENDPOINTS:
//   POST /api/patients/{id}/anamnesis - Criar anamnese para paciente
//   GET  /api/patients/{id}/anamnesis - Listar anamneses do paciente
//   GET  /api/anamnesis/{id}          - Buscar anamnese por ID
//   PUT  /api/anamnesis/{id}          - Atualizar anamnese
//   DELETE /api/anamnesis/{id}        - Remover anamnese
//   GET  /api/anamnesis               - Listar todas anamneses (filtros)
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
// STRUCT ANAMNESEHANDLER
// ================================================================
type AnamneseHandler struct {
	anamneseService *services.AnamneseService
	validator       *validator.Validate
}

// ================================================================
// FUNÇÃO NEWANAMNESEHANDLER()
// ================================================================
func NewAnamneseHandler(anamneseService *services.AnamneseService) *AnamneseHandler {
	return &AnamneseHandler{
		anamneseService: anamneseService,
		validator:       validator.New(),
	}
}

// ================================================================
// HANDLER: CREATE
// ================================================================
// Endpoint: POST /api/patients/{id}/anamnesis
func (h *AnamneseHandler) Create(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair ID do paciente ---
	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	// --- 2. Obter ID do usuário responsável ---
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID do usuário")
		return
	}

	// --- 3. Decodificar JSON ---
	var req services.CreateAnamneseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// --- 4. Validar dados ---
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	// --- 5. Chamar serviço ---
	anamnese, err := h.anamneseService.Create(patientID, userID, req)
	if err != nil {
		log.Printf("❌ Erro ao criar anamnese: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, anamnese)
}

// ================================================================
// HANDLER: GET BY PATIENT
// ================================================================
// Endpoint: GET /api/patients/{id}/anamnesis
func (h *AnamneseHandler) GetByPatient(w http.ResponseWriter, r *http.Request) {
	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	anamneses, err := h.anamneseService.GetByPatient(patientID)
	if err != nil {
		log.Printf("❌ Erro ao listar anamneses: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar anamneses")
		return
	}

	utils.SendSuccess(w, http.StatusOK, anamneses)
}

// ================================================================
// HANDLER: GET BY ID
// ================================================================
// Endpoint: GET /api/anamnesis/{id}
func (h *AnamneseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	anamnese, err := h.anamneseService.GetByID(id)
	if err != nil {
		if err.Error() == "anamnese não encontrada" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar anamnese: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar anamnese")
		return
	}

	utils.SendSuccess(w, http.StatusOK, anamnese)
}

// ================================================================
// HANDLER: LIST
// ================================================================
// Endpoint: GET /api/anamnesis
func (h *AnamneseHandler) List(w http.ResponseWriter, r *http.Request) {
	req := services.ListAnamneseRequest{
		PatientID: r.URL.Query().Get("patient_id"),
		Type:      r.URL.Query().Get("type"),
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

	anamneses, total, err := h.anamneseService.List(req)
	if err != nil {
		log.Printf("❌ Erro ao listar anamneses: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar anamneses")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": anamneses,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// ================================================================
// HANDLER: UPDATE
// ================================================================
// Endpoint: PUT /api/anamnesis/{id}
func (h *AnamneseHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req services.UpdateAnamneseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	anamnese, err := h.anamneseService.Update(id, req)
	if err != nil {
		if err.Error() == "anamnese não encontrada" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar anamnese: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, anamnese)
}

// ================================================================
// HANDLER: DELETE
// ================================================================
// Endpoint: DELETE /api/anamnesis/{id}
func (h *AnamneseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.anamneseService.Delete(id); err != nil {
		if err.Error() == "anamnese não encontrada" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao deletar anamnese: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao deletar anamnese")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Anamnese removida com sucesso",
	})
}
