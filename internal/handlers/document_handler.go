// ================================================================
// PACOTE HANDLERS - DOCUMENT HANDLER
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
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	// --- 1. Extrair ID do paciente da URL ---
	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	// --- 2. Verificar se o paciente existe ---
	// (será verificado no service)

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

	// --- 6. Chamar serviço para upload ---
	document, err := h.documentService.Upload(patientID, documentType, file, fileHeader, userID)
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
	patientIDStr := chi.URLParam(r, "id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do paciente inválido")
		return
	}

	documents, err := h.documentService.GetByPatient(patientID)
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
	docIDStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do documento inválido")
		return
	}

	document, err := h.documentService.GetByID(docID)
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
	docIDStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do documento inválido")
		return
	}

	// Buscar documento
	document, err := h.documentService.GetByID(docID)
	if err != nil {
		if err.Error() == "documento não encontrado" {
			utils.SendError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("❌ Erro ao buscar documento: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao buscar documento")
		return
	}

	// Construir caminho do arquivo
	filename := filepath.Base(document.FileURL)
	fullPath := filepath.Join("uploads/documents", filename)

	// Verificar se o arquivo existe
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		utils.SendError(w, http.StatusNotFound, "arquivo não encontrado no servidor")
		return
	}

	// Abrir arquivo
	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("❌ Erro ao abrir arquivo: %v", err)
		utils.SendError(w, http.StatusInternalServerError, "erro ao abrir arquivo")
		return
	}
	defer file.Close()

	// Configurar headers para download
	w.Header().Set("Content-Type", document.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+document.FileName+"\"")
	w.Header().Set("Content-Length", string(rune(document.FileSize)))

	// Enviar arquivo
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("❌ Erro ao enviar arquivo: %v", err)
	}
}

// ================================================================
// HANDLER: UPDATE STATUS
// ================================================================
// Endpoint: PATCH /api/documents/{id}/status
//
// Aprova ou rejeita um documento
func (h *DocumentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	document, err := h.documentService.UpdateStatus(docID, req, userID)
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
	docIDStr := chi.URLParam(r, "id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "ID do documento inválido")
		return
	}

	if err := h.documentService.Delete(docID); err != nil {
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
