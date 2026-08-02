// ================================================================
// PACOTE SERVICES - DOCUMENT SERVICE
// ================================================================
// Camada de serviço responsável pela gestão de documentos dos pacientes.
//
// RESPONSABILIDADES:
// 1. Upload de documentos
// 2. Listagem de documentos por paciente
// 3. Download de documentos
// 4. Aprovação/Rejeição de documentos
// 5. Remoção de documentos
//
// TIPOS DE DOCUMENTOS:
//   - rg_cpf: RG ou CPF
//   - comprovante_residencia: Comprovante de residência
//   - laudo_medico: Laudo médico
//   - receita_medica: Receita médica
//   - autorizacao_anvisa: Autorização da ANVISA
//   - termo_consentimento: Termo de consentimento
// ================================================================

package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cannacare-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT DOCUMENTSERVICE
// ================================================================
type DocumentService struct {
	db           *gorm.DB
	uploadPath   string   // Caminho onde os arquivos serão salvos
	maxFileSize  int64    // Tamanho máximo do arquivo em bytes
	allowedTypes []string // Tipos de arquivo permitidos
}

// ================================================================
// FUNÇÃO NEWDOCUMENTSERVICE()
// ================================================================
func NewDocumentService(db *gorm.DB) *DocumentService {
	return &DocumentService{
		db:          db,
		uploadPath:  "uploads/documents",
		maxFileSize: 10 * 1024 * 1024, // 10 MB
		allowedTypes: []string{
			"application/pdf",
			"image/jpeg",
			"image/png",
			"image/jpg",
		},
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

// DocumentResponse - Resposta com dados do documento
type DocumentResponse struct {
	ID           string `json:"id"`
	PatientID    string `json:"patient_id"`
	DocumentType string `json:"document_type"`
	FileName     string `json:"file_name"`
	FileURL      string `json:"file_url"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// UpdateDocumentStatusRequest - Para aprovar/rejeitar documento
type UpdateDocumentStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=aprovado rejeitado"`
	Reason string `json:"reason" validate:"omitempty"`
}

// ================================================================
// FUNÇÃO UPLOAD - CORRIGIDA COM ASSOCIATION_ID
// ================================================================
func (s *DocumentService) Upload(associationID uuid.UUID, patientID uuid.UUID, documentType string, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID) (*DocumentResponse, error) {
	// --- 1. Validar tipo de documento ---
	if !s.isValidDocumentType(documentType) {
		return nil, fmt.Errorf("tipo de documento inválido: %s", documentType)
	}

	// --- 2. Validar tamanho do arquivo ---
	if fileHeader.Size > s.maxFileSize {
		return nil, fmt.Errorf("arquivo muito grande: %d bytes (máximo: %d MB)", fileHeader.Size, s.maxFileSize/1024/1024)
	}

	// --- 3. Validar tipo do arquivo (MIME) ---
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return nil, err
	}
	mimeType := http.DetectContentType(buffer)
	if !s.isAllowedMimeType(mimeType) {
		return nil, fmt.Errorf("tipo de arquivo não permitido: %s", mimeType)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	// --- 4. Verificar se o paciente existe E pertence à associação ---
	var patient models.Patient
	if err := s.db.Where("id = ? AND association_id = ?", patientID, associationID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado ou não pertence à sua associação")
		}
		return nil, err
	}

	// --- 5. Gerar nome único para o arquivo ---
	extension := filepath.Ext(fileHeader.Filename)
	if extension == "" {
		ext := getExtensionFromMime(mimeType)
		extension = "." + ext
	}
	filename := fmt.Sprintf("%s_%s_%d%s",
		patientID.String(),
		documentType,
		time.Now().Unix(),
		extension,
	)
	fullPath := filepath.Join(s.uploadPath, filename)

	// --- 6. Salvar o arquivo no disco ---
	if err := os.MkdirAll(s.uploadPath, 0755); err != nil {
		return nil, err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	// --- 7. Criar registro no banco (COM association_id) ---
	document := &models.PatientDocument{
		AssociationID: associationID, // ← ⚠️ ESSENCIAL!
		PatientID:     patientID,
		DocumentType:  documentType,
		FileURL:       fmt.Sprintf("/uploads/documents/%s", filename),
		FileName:      fileHeader.Filename,
		FileSize:      fileHeader.Size,
		MimeType:      mimeType,
		Status:        "em_analise",
	}

	if err := s.db.Create(document).Error; err != nil {
		os.Remove(fullPath)
		return nil, err
	}

	return toDocumentResponse(document), nil
}

// ================================================================
// FUNÇÃO GETBYPATIENT()
// ================================================================
// Lista todos os documentos de um paciente
func (s *DocumentService) GetByPatient(patientID uuid.UUID) ([]DocumentResponse, error) {
	var documents []models.PatientDocument

	if err := s.db.Where("patient_id = ?", patientID).Order("created_at DESC").Find(&documents).Error; err != nil {
		return nil, err
	}

	var responses []DocumentResponse
	for _, doc := range documents {
		responses = append(responses, *toDocumentResponse(&doc))
	}

	return responses, nil
}

// ================================================================
// FUNÇÃO GETBYID()
// ================================================================
// Busca um documento pelo ID
func (s *DocumentService) GetByID(id uuid.UUID) (*DocumentResponse, error) {
	var document models.PatientDocument
	if err := s.db.Where("id = ?", id).First(&document).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("documento não encontrado")
		}
		return nil, err
	}
	return toDocumentResponse(&document), nil
}

// ================================================================
// FUNÇÃO UPDATESTATUS()
// ================================================================
// Aprova ou rejeita um documento
func (s *DocumentService) UpdateStatus(id uuid.UUID, req UpdateDocumentStatusRequest, userID uuid.UUID) (*DocumentResponse, error) {
	var document models.PatientDocument
	if err := s.db.Where("id = ?", id).First(&document).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("documento não encontrado")
		}
		return nil, err
	}

	// Atualizar status
	document.Status = req.Status
	document.ReviewedBy = &userID
	now := time.Now().Format("2006-01-02 15:04:05")
	document.ReviewedAt = &now

	if err := s.db.Save(&document).Error; err != nil {
		return nil, err
	}

	return toDocumentResponse(&document), nil
}

// ================================================================
// FUNÇÃO DELETE()
// ================================================================
// Remove um documento (físico e do banco)
func (s *DocumentService) Delete(id uuid.UUID) error {
	var document models.PatientDocument
	if err := s.db.Where("id = ?", id).First(&document).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("documento não encontrado")
		}
		return err
	}

	// Remover arquivo físico
	filename := filepath.Base(document.FileURL)
	fullPath := filepath.Join(s.uploadPath, filename)
	if err := os.Remove(fullPath); err != nil {
		// Se não conseguir remover, apenas loga o erro
		fmt.Printf("⚠️ Erro ao remover arquivo: %v\n", err)
	}

	// Remover registro do banco
	if err := s.db.Delete(&document).Error; err != nil {
		return err
	}

	return nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// toDocumentResponse - Converte models.PatientDocument para DocumentResponse
func toDocumentResponse(document *models.PatientDocument) *DocumentResponse {
	return &DocumentResponse{
		ID:           document.ID.String(),
		PatientID:    document.PatientID.String(),
		DocumentType: document.DocumentType,
		FileName:     document.FileName,
		FileURL:      document.FileURL,
		FileSize:     document.FileSize,
		MimeType:     document.MimeType,
		Status:       document.Status,
		CreatedAt:    document.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    document.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// isValidDocumentType - Valida se o tipo de documento é permitido
func (s *DocumentService) isValidDocumentType(docType string) bool {
	validTypes := []string{
		"rg_cpf",
		"comprovante_residencia",
		"laudo_medico",
		"receita_medica",
		"autorizacao_anvisa",
		"termo_consentimento",
	}
	for _, t := range validTypes {
		if t == docType {
			return true
		}
	}
	return false
}

// isAllowedMimeType - Valida se o MIME type é permitido
func (s *DocumentService) isAllowedMimeType(mimeType string) bool {
	for _, t := range s.allowedTypes {
		if t == mimeType {
			return true
		}
	}
	return false
}

// getExtensionFromMime - Retorna a extensão do arquivo baseado no MIME type
func getExtensionFromMime(mimeType string) string {
	extensions := map[string]string{
		"application/pdf": "pdf",
		"image/jpeg":      "jpg",
		"image/jpg":       "jpg",
		"image/png":       "png",
	}
	if ext, ok := extensions[mimeType]; ok {
		return ext
	}
	return "bin"
}
