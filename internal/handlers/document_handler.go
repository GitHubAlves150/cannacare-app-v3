// ================================================================
// PACOTE HANDLERS - DOCUMENT HANDLER (COM MULTI-TENANCY)
// ================================================================
// Camada HTTP que lida com as requisições de documentos.
//
// ENDPOINTS:
//   POST /api/patients/{id}/documents - Upload de documento
//   GET  /api/patients/{id}/documents - Listar documentos do paciente
//   GET  /api/documents/{id}           - Download de documento
//   PATCH /api/documents/{id}/status   - Aprovar/Rejeitar documento
//   DELETE /api/documents/{id}         - Remover documento
// ================================================================

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"cannacare-backend/internal/middleware"
	"cannacare-backend/internal/services"
	"cannacare-backend/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ================================================================
// STRUCT DOCUMENTHANDLER
// ================================================================
type DocumentHandler struct {
	documentService *services.DocumentService
	validator       *validator.Validate
}

// ================================================================
// FUNÇÃO NEWDOCUMENTHANDLER()
// ================================================================
func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		validator:       validator.New(),
	}
}

// ================================================================
// FUNÇÃO AUXILIAR: extractAssociationID
// ================================================================
func (h *DocumentHandler) extractAssociationID(r *http.Request) (uuid.UUID, error) {
	associationIDStr, _ := r.Context().Value(middleware.AssociationIDKey).(string)
	if associationIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(associationIDStr)
}

// ================================================================
// HANDLER: UPLOAD
// ================================================================
// Endpoint: POST /api/patients/{id}/documents
//
// # Upload de um documento para o paciente
//
// EXEMPLO DE REQUISIÇÃO (multipart/form-data):
//   - document_type: rg_cpf
//   - file: arquivo.pdf ou imagem.jpg
func (h *DocumentHandler) Upload(w http.ResponseWriter, r *http.Request) {
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

	// --- 2. Extrair ID do paciente da URL ---
	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	// --- 3. Extrair document_type do form ---
	documentType := r.FormValue("document_type")
	if documentType == "" {
		utils.SendError(w, http.StatusBadRequest, "campo document_type é obrigatório")
		return
	}

	// --- 4. Obter o arquivo do form ---
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		if err == http.ErrMissingFile {
			utils.SendError(w, http.StatusBadRequest, "arquivo é obrigatório")
			return
		}
		utils.SendError(w, http.StatusBadRequest, "erro ao ler arquivo: "+err.Error())
		return
	}
	defer file.Close()

	// --- 5. Obter ID do usuário que está fazendo o upload ---
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID do usuário")
		return
	}

	// --- 6. Chamar serviço para upload (passando association_id) ---
	document, err := h.documentService.Upload(associationID, patientID, documentType, file, fileHeader, userID)
	if err != nil {
		log.Printf("❌ Erro ao fazer upload: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// --- 7. Retornar sucesso ---
	utils.SendSuccess(w, http.StatusCreated, document)
}

// ================================================================
// HANDLER: LIST BY PATIENT
// ================================================================
// Endpoint: GET /api/patients/{id}/documents
//
// Lista todos os documentos de um paciente
func (h *DocumentHandler) ListByPatient(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	documents, err := h.documentService.GetByPatient(associationID, patientID)
	if err != nil {
		log.Printf("❌ Erro ao listar documentos: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao listar documentos")
		return
	}

	utils.SendSuccess(w, http.StatusOK, documents)
}

// ================================================================
// HANDLER: GET BY ID
// ================================================================
// Endpoint: GET /api/documents/{id}
//
// Retorna os dados de um documento específico
func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	docIDStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do documento inválido")
		return
	}

	document, err := h.documentService.GetByID(associationID, docID)
	if err != nil {
		if err.Error() == "documento não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar documento: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar documento")
		return
	}

	utils.SendSuccess(w, http.StatusOK, document)
}

// ================================================================
// HANDLER: DOWNLOAD
// ================================================================
// Endpoint: GET /api/documents/{id}/download
//
// Faz o download do arquivo do documento
func (h *DocumentHandler) Download(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	docIDStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do documento inválido")
		return
	}

	// Gera uma URL assinada do S3, válida por 15 minutos, e devolve em
	// JSON (em vez de redirecionar) — assim o front só abre a URL numa
	// aba nova, sem precisar lidar com CORS entre domínios.
	downloadURL, fileName, err := h.documentService.GetDownloadURL(associationID, docID)
	if err != nil {
		if err.Error() == "documento não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao gerar link de download: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao gerar link de download")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"url":       downloadURL,
		"file_name": fileName,
	})
}

// ================================================================
// HANDLER: UPDATE STATUS
// ================================================================
// Endpoint: PATCH /api/documents/{id}/status
//
// Aprova ou rejeita um documento
func (h *DocumentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	docIDStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do documento inválido")
		return
	}

	var req services.UpdateDocumentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	// Obter ID do usuário
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID do usuário")
		return
	}

	document, err := h.documentService.UpdateStatus(associationID, docID, req, userID)
	if err != nil {
		if err.Error() == "documento não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao atualizar status do documento: %v", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, document)
}

// ================================================================
// HANDLER: DELETE
// ================================================================
// Endpoint: DELETE /api/documents/{id}
//
// Remove um documento
func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	associationID, err := h.extractAssociationID(r)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "erro ao obter ID da associação")
		return
	}
	if associationID == uuid.Nil {
		utils.SendError(w, http.StatusUnauthorized, "associação não identificada")
		return
	}

	docIDStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do documento inválido")
		return
	}

	if err := h.documentService.Delete(associationID, docID); err != nil {
		if err.Error() == "documento não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao deletar documento: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao deletar documento")
		return
	}

	utils.SendSuccess(w, http.StatusOK, map[string]string{
		"message": "Documento removido com sucesso",
	})
}