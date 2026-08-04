// ================================================================
// PACOTE HANDLERS - DOCTOR HANDLER (COM MULTI-TENANCY COMPLETO)
// ================================================================
// Camada HTTP que lida com as requisições de médicos.
//
// ENDPOINTS:
//   POST   /api/doctors           - Criar médico
//   GET    /api/doctors           - Listar médicos (com filtros)
//   GET    /api/doctors/:id       - Buscar médico por ID
//   PUT    /api/doctors/:id       - Atualizar médico
//   DELETE /api/doctors/:id       - Remover médico (soft delete)
//   GET    /api/doctors/top       - Médicos que mais prescrevem
// ================================================================

package handlers

import (
	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ================================================================
// STRUCT DOCTORHANDLER
// ================================================================
type DoctorHandler struct {
	doctorService *services.DoctorService
	validator     *validator.Validate
}

// ================================================================
// FUNÇÃO NEWDOCTORHANDLER()
// ================================================================
func NewDoctorHandler(doctorService *services.DoctorService) *DoctorHandler {
	return &DoctorHandler{
		doctorService: doctorService,
		validator:     validator.New(),
	}
}

// ================================================================
// FUNÇÃO AUXILIAR: extractAssociationID
// ================================================================
func (h *DoctorHandler) extractAssociationID(r *http.Request) (uuid.UUID, error) {
	associationIDStr := r.Context().Value(middleware.AssociationIDKey).(string)
	if associationIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(associationIDStr)
}

// ================================================================
// HANDLER: CREATE DOCTOR
// ================================================================
func (h *DoctorHandler) Create(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	// --- 2. Decodificar JSON ---
	var req services.CreateDoctorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// --- 3. Validar dados ---
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	// --- 4. Chamar serviço (passando association_id) ---
	doctor, err := h.doctorService.Create(associationID, req)
	if err != nil {
		log.Printf("❌ Erro ao criar médico: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, doctor)
}

// ================================================================
// HANDLER: GET DOCTOR BY ID (CORRIGIDO)
// ================================================================
func (h *DoctorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	// --- 2. Extrair ID da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- 3. Chamar serviço (passando association_id) ---
	doctor, err := h.doctorService.GetByID(associationID, id)
	if err != nil {
		if err.Error() == "médico não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar médico: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar médico")
		return
	}

	utils.SendSuccess(w, http.StatusOK, doctor)
}

// ================================================================
// HANDLER: LIST DOCTORS (CORRIGIDO)
// ================================================================
func (h *DoctorHandler) List(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	// --- 2. Extrair filtros da URL ---
	req := services.ListDoctorRequest{
		Name:      r.URL.Query().Get("name"),
		CRM:       r.URL.Query().Get("crm"),
		Specialty: r.URL.Query().Get("specialty"),
	}

	if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			req.IsActive = &isActive
		}
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, _ := strconv.Atoi(pageStr)
		req.Page = page
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		req.Limit = limit
	}

	// --- 3. Chamar serviço (passando association_id) ---
	doctors, total, err := h.doctorService.List(associationID, req)
	if err != nil {
		log.Printf("❌ Erro ao listar médicos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar médicos")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": doctors,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// ================================================================
// HANDLER: UPDATE DOCTOR (CORRIGIDO)
// ================================================================
func (h *DoctorHandler) Update(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	// --- 2. Extrair ID da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- 3. Decodificar JSON ---
	var req services.UpdateDoctorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// --- 4. Chamar serviço (passando association_id) ---
	doctor, err := h.doctorService.Update(associationID, id, req)
	if err != nil {
		if err.Error() == "médico não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar médico: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, doctor)
}

// ================================================================
// HANDLER: DELETE DOCTOR (CORRIGIDO)
// ================================================================
func (h *DoctorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	// --- 2. Extrair ID da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- 3. Chamar serviço (passando association_id) ---
	if err := h.doctorService.Delete(associationID, id); err != nil {
		if err.Error() == "médico não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao deletar médico: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao deletar médico")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Médico removido com sucesso",
	})
}

// ================================================================
// HANDLER: GET TOP DOCTORS (CORRIGIDO)
// ================================================================
func (h *DoctorHandler) GetTopDoctors(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	// --- 2. Chamar serviço (passando association_id) ---
	doctors, err := h.doctorService.GetTopDoctors(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar top médicos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar top médicos")
		return
	}

	utils.SendSuccess(w, http.StatusOK, doctors)
}
