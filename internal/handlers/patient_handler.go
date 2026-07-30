// ================================================================
// CANNACARE - PATIENT HANDLER (COMPLETO)
// ================================================================
// Camada HTTP que lida com as requisições de pacientes.
//
// ENDPOINTS:
//   POST   /api/patients           - Criar paciente
//   GET    /api/patients           - Listar pacientes (com filtro)
//   GET    /api/patients/{id}      - Buscar paciente por ID
//   PUT    /api/patients/{id}      - Atualizar paciente
//   PATCH  /api/patients/{id}/status - Mudar status do paciente
//   DELETE /api/patients/{id}      - Remover paciente (soft delete)
//   GET    /api/patients/stats     - Estatísticas de pacientes
//
// MULTI-TENANCY:
//   TODAS as operações extraem association_id do Context (JWT)
//   e passam para os services.
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
// FUNÇÃO AUXILIAR: extractAssociationID
// ================================================================
// Extrai o association_id do Context e retorna como UUID.
// Esta função é usada por TODOS os métodos para evitar repetição.
// ================================================================
func (h *PatientHandler) extractAssociationID(r *http.Request) (uuid.UUID, error) {
	associationIDStr := r.Context().Value(middleware.AssociationIDKey).(string)
	if associationIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(associationIDStr)
}

// ================================================================
// HANDLER: CREATE - Criar paciente
// ================================================================
// Endpoint: POST /api/patients
//
// FLUXO:
//   1. Extrai association_id do Context (definido pelo AuthMiddleware)
//   2. Decodifica o JSON da requisição
//   3. Valida os dados
//   4. Chama o serviço para criar o paciente
//   5. Retorna o paciente criado
// ================================================================
func (h *PatientHandler) Create(w http.ResponseWriter, r *http.Request) {
	// --- PASSO 1: Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}

	// --- PASSO 2: Decodificar JSON ---
	var req services.CreatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// --- PASSO 3: Validar dados ---
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	// --- PASSO 4: Chamar serviço (passando association_id) ---
	patient, err := h.patientService.Create(associationID, req)
	if err != nil {
		log.Printf("❌ Erro ao criar paciente: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, patient)
}

// ================================================================
// HANDLER: GETBYID - Buscar paciente por ID
// ================================================================
// Endpoint: GET /api/patients/{id}
//
// ⚠️ IMPORTANTE: SEMPRE filtra por association_id!
// Nunca permita que um usuário veja um paciente de outra associação.
// ================================================================
func (h *PatientHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// --- PASSO 1: Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}

	// --- PASSO 2: Extrair ID do paciente da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- PASSO 3: Chamar serviço (passando association_id) ---
	patient, err := h.patientService.GetByID(associationID, id)
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
// HANDLER: LIST - Listar pacientes (com filtro por associação)
// ================================================================
// Endpoint: GET /api/patients
//
// FLUXO:
//   1. Extrai association_id do Context
//   2. Aplica filtros (nome, CPF, status, etc)
//   3. Chama o serviço com association_id + filtros
//   4. Retorna a lista de pacientes (SÓ DA ASSOCIAÇÃO DO USUÁRIO)
// ================================================================
func (h *PatientHandler) List(w http.ResponseWriter, r *http.Request) {
	// --- PASSO 1: Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}

	// --- PASSO 2: Extrair filtros da URL ---
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

	// --- PASSO 3: Chamar serviço (passando association_id) ---
	patients, total, err := h.patientService.List(associationID, req)
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
// HANDLER: UPDATE - Atualizar paciente
// ================================================================
// Endpoint: PUT /api/patients/{id}
// ================================================================
func (h *PatientHandler) Update(w http.ResponseWriter, r *http.Request) {
	// --- PASSO 1: Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}

	// --- PASSO 2: Extrair ID do paciente da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- PASSO 3: Decodificar JSON ---
	var req services.UpdatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// --- PASSO 4: Chamar serviço (passando association_id) ---
	patient, err := h.patientService.Update(associationID, id, req)
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
// HANDLER: DELETE - Remover paciente (soft delete)
// ================================================================
// Endpoint: DELETE /api/patients/{id}
// ================================================================
func (h *PatientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// --- PASSO 1: Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}

	// --- PASSO 2: Extrair ID do paciente da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- PASSO 3: Chamar serviço (passando association_id) ---
	if err := h.patientService.Delete(associationID, id); err != nil {
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
// HANDLER: UPDATESTATUS - Mudar status do paciente
// ================================================================
// Endpoint: PATCH /api/patients/{id}/status
// ================================================================
func (h *PatientHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	// --- PASSO 1: Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}

	// --- PASSO 2: Extrair ID do paciente da URL ---
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// --- PASSO 3: Decodificar JSON ---
	var req services.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// --- PASSO 4: Validar status ---
	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "status inválido: "+err.Error())
		return
	}

	// --- PASSO 5: Obter ID do usuário que está fazendo a alteração ---
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID do usuário")
		return
	}

	// --- PASSO 6: Chamar serviço (passando association_id) ---
	patient, err := h.patientService.UpdateStatus(associationID, id, req, userID)
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
// HANDLER: GETSTATISTICS - Estatísticas de pacientes
// ================================================================
// Endpoint: GET /api/patients/stats
// ================================================================
func (h *PatientHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	// --- PASSO 1: Extrair association_id do Context ---
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}

	// --- PASSO 2: Chamar serviço (passando association_id) ---
	stats, err := h.patientService.GetStatistics(associationID)
	if err != nil {
		log.Printf("❌ Erro ao buscar estatísticas: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar estatísticas")
		return
	}

	utils.SendSuccess(w, http.StatusOK, stats)
}