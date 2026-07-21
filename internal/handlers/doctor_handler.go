// ================================================================
// PACOTE HANDLERS - DOCTOR HANDLER
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
// HANDLER: CREATE DOCTOR
// ================================================================
// Endpoint: POST /api/doctors
//
// Cria um novo médico no sistema
func (h *DoctorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req services.CreateDoctorRequest

	// --- 1. Decodificar JSON ---
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// --- 2. Validar dados ---
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	// --- 3. Chamar serviço ---
	doctor, err := h.doctorService.Create(req)
	if err != nil {
		log.Printf("❌ Erro ao criar médico: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// --- 4. Retornar sucesso ---
	utils.SendSuccess(w, http.StatusCreated, doctor)
}

// ================================================================
// HANDLER: GET DOCTOR BY ID
// ================================================================
// Endpoint: GET /api/doctors/:id
//
// Busca um médico pelo ID
func (h *DoctorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair ID da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- 2. Buscar médico ---
	doctor, err := h.doctorService.GetByID(id)
	if err != nil {
		if err.Error() == "médico não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar médico: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar médico")
		return
	}

	// --- 3. Retornar sucesso ---
	utils.SendSuccess(w, http.StatusOK, doctor)
}

// ================================================================
// HANDLER: LIST DOCTORS
// ================================================================
// Endpoint: GET /api/doctors
//
// Lista médicos com filtros e paginação
func (h *DoctorHandler) List(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair query params ---
	req := services.ListDoctorRequest{
		Name:      r.URL.Query().Get("name"),
		CRM:       r.URL.Query().Get("crm"),
		Specialty: r.URL.Query().Get("specialty"),
	}

	// Parse is_active
	if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			req.IsActive = &isActive
		}
	}

	// Parse paginação
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

	// --- 2. Chamar serviço ---
	doctors, total, err := h.doctorService.List(req)
	if err != nil {
		log.Printf("❌ Erro ao listar médicos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar médicos")
		return
	}

	// --- 3. Retornar sucesso com metadados ---
	utils.SendSuccess(w, http.StatusOK, map[string]interface{}{
		"items": doctors,
		"total": total,
		"page":  req.Page,
		"limit": req.Limit,
	})
}

// ================================================================
// HANDLER: UPDATE DOCTOR
// ================================================================
// Endpoint: PUT /api/doctors/:id
//
// Atualiza os dados de um médico
func (h *DoctorHandler) Update(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair ID da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- 2. Decodificar JSON ---
	var req services.UpdateDoctorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// --- 3. Chamar serviço ---
	doctor, err := h.doctorService.Update(id, req)
	if err != nil {
		if err.Error() == "médico não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar médico: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// --- 4. Retornar sucesso ---
	utils.SendSuccess(w, http.StatusOK, doctor)
}

// ================================================================
// HANDLER: DELETE DOCTOR
// ================================================================
// Endpoint: DELETE /api/doctors/:id
//
// Remove um médico (soft delete)
func (h *DoctorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// --- 1. Extrair ID da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- 2. Chamar serviço ---
	if err := h.doctorService.Delete(id); err != nil {
		if err.Error() == "médico não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao deletar médico: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao deletar médico")
		return
	}

	// --- 3. Retornar sucesso ---
	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Médico removido com sucesso",
	})
}

// ================================================================
// HANDLER: GET TOP DOCTORS
// ================================================================
// Endpoint: GET /api/doctors/top
//
// Retorna os médicos que mais prescrevem
func (h *DoctorHandler) GetTopDoctors(w http.ResponseWriter, r *http.Request) {
	doctors, err := h.doctorService.GetTopDoctors()
	if err != nil {
		log.Printf("❌ Erro ao buscar top médicos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar top médicos")
		return
	}

	utils.SendSuccess(w, http.StatusOK, doctors)
}
