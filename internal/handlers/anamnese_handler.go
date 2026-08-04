// ================================================================
// PACOTE HANDLERS - ANAMNESE HANDLER
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

type AnamneseHandler struct {
	anamneseService *services.AnamneseService
	validator       *validator.Validate
}

func NewAnamneseHandler(anamneseService *services.AnamneseService) *AnamneseHandler {
	return &AnamneseHandler{
		anamneseService: anamneseService,
		validator:       validator.New(),
	}
}

func (h *AnamneseHandler) extractAssociationID(r *http.Request) (uuid.UUID, error) {
	associationIDStr, _ := r.Context().Value(middleware.AssociationIDKey).(string)
	if associationIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(associationIDStr)
}

// POST /api/patients/{id}/anamnesis
func (h *AnamneseHandler) Create(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID do usuário")
		return
	}

	var req services.CreateAnamneseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	anamnese, err := h.anamneseService.Create(associationID, patientID, userID, req)
	if err != nil {
		log.Printf("❌ Erro ao criar anamnese: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, anamnese)
}

// GET /api/patients/{id}/anamnesis
func (h *AnamneseHandler) GetByPatient(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	anamneses, err := h.anamneseService.GetByPatient(associationID, patientID)
	if err != nil {
		log.Printf("❌ Erro ao listar anamneses: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar anamneses")
		return
	}

	utils.SendSuccess(w, http.StatusOK, anamneses)
}

// GET /api/anamnesis/{id}
func (h *AnamneseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	anamnese, err := h.anamneseService.GetByID(associationID, id)
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

// GET /api/anamnesis
func (h *AnamneseHandler) List(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

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

	anamneses, total, err := h.anamneseService.List(associationID, req)
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

// PUT /api/anamnesis/{id}
func (h *AnamneseHandler) Update(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

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

	anamnese, err := h.anamneseService.Update(associationID, id, req)
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

// DELETE /api/anamnesis/{id}
func (h *AnamneseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil || associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.anamneseService.Delete(associationID, id); err != nil {
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