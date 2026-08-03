// ================================================================
// PACOTE SERVICES - DOCUMENT SERVICE (COM AWS S3)
// ================================================================
// ⚠️ MIGRADO: antes salvava em disco local (uploads/documents),
// o que perde os arquivos a cada reinício/deploy em produção.
// Agora salva no S3 — os arquivos sobrevivem independente do que
// acontecer com o servidor.
//
// SEGURANÇA: o bucket é PRIVADO (documentos de paciente são dado
// sensível). Download não serve o arquivo direto — gera uma URL
// assinada (presigned) que expira em poucos minutos.
//
// VARIÁVEIS DE AMBIENTE NECESSÁRIAS:
//   AWS_ACCESS_KEY_ID
//   AWS_SECRET_ACCESS_KEY
//   AWS_REGION       (ex: sa-east-1)
//   AWS_S3_BUCKET    (ex: cannacare-documents)
// ================================================================

package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cannacare-backend/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT DOCUMENTSERVICE
// ================================================================
type DocumentService struct {
	db            *gorm.DB
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	maxFileSize   int64
	allowedTypes  []string
}

// ================================================================
// FUNÇÃO NEWDOCUMENTSERVICE()
// ================================================================
func NewDocumentService(db *gorm.DB) *DocumentService {
	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("AWS_S3_BUCKET")

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		// Não derruba o servidor inteiro por causa disso — só loga.
		// Upload/download vão falhar com mensagem clara até configurar.
		fmt.Printf("⚠️ erro ao carregar configuração da AWS: %v\n", err)
	}

	client := s3.NewFromConfig(cfg)

	return &DocumentService{
		db:            db,
		s3Client:      client,
		presignClient: s3.NewPresignClient(client),
		bucket:        bucket,
		maxFileSize:   10 * 1024 * 1024, // 10 MB
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

type DocumentResponse struct {
	ID           string `json:"id"`
	PatientID    string `json:"patient_id"`
	DocumentType string `json:"document_type"`
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type UpdateDocumentStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=aprovado rejeitado"`
	Reason string `json:"reason" validate:"omitempty"`
}

// ================================================================
// FUNÇÃO UPLOAD()
// ================================================================
func (s *DocumentService) Upload(associationID, patientID uuid.UUID, documentType string, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID) (*DocumentResponse, error) {
	if s.bucket == "" {
		return nil, errors.New("armazenamento de arquivos não configurado (AWS_S3_BUCKET ausente)")
	}

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
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}

	// --- 5. Montar a chave (key) do objeto no S3 ---
	// Formato: documents/{association_id}/{patient_id}/{tipo}_{timestamp}{extensão}
	// Isso já isola fisicamente os arquivos por associação dentro do bucket.
	extension := filepath.Ext(fileHeader.Filename)
	if extension == "" {
		extension = "." + getExtensionFromMime(mimeType)
	}
	key := fmt.Sprintf("documents/%s/%s/%s_%d%s",
		associationID.String(),
		patientID.String(),
		documentType,
		time.Now().Unix(),
		extension,
	)

	// --- 6. Enviar para o S3 ---
	_, err := s.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(mimeType),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar arquivo para o armazenamento: %w", err)
	}

	// --- 7. Criar registro no banco ---
	// FileURL guarda a CHAVE do objeto no S3, não uma URL fixa — o
	// bucket é privado, então a URL de verdade é gerada na hora do
	// download (presigned, expira em poucos minutos).
	document := &models.PatientDocument{
		AssociationID: associationID,
		PatientID:     patientID,
		DocumentType:  documentType,
		FileURL:       key,
		FileName:      fileHeader.Filename,
		FileSize:      fileHeader.Size,
		MimeType:      mimeType,
		Status:        "em_analise",
	}

	if err := s.db.Create(document).Error; err != nil {
		// Se falhar ao salvar no banco, remove o que já foi enviado ao S3
		_, _ = s.s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		})
		return nil, err
	}

	return toDocumentResponse(document), nil
}

// ================================================================
// FUNÇÃO GETDOWNLOADURL() - gera URL assinada temporária
// ================================================================
// Válida por 15 minutos. Nunca expõe o bucket publicamente.
func (s *DocumentService) GetDownloadURL(associationID, id uuid.UUID) (string, string, error) {
	var document models.PatientDocument
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&document).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", errors.New("documento não encontrado")
		}
		return "", "", err
	}

	presignedReq, err := s.presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(document.FileURL),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", "", fmt.Errorf("erro ao gerar link de download: %w", err)
	}

	return presignedReq.URL, document.FileName, nil
}

// ================================================================
// FUNÇÃO GETBYPATIENT()
// ================================================================
func (s *DocumentService) GetByPatient(associationID, patientID uuid.UUID) ([]DocumentResponse, error) {
	var documents []models.PatientDocument

	if err := s.db.Where("patient_id = ? AND association_id = ?", patientID, associationID).
		Order("created_at DESC").Find(&documents).Error; err != nil {
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
func (s *DocumentService) GetByID(associationID, id uuid.UUID) (*DocumentResponse, error) {
	var document models.PatientDocument
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&document).Error; err != nil {
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
func (s *DocumentService) UpdateStatus(associationID, id uuid.UUID, req UpdateDocumentStatusRequest, userID uuid.UUID) (*DocumentResponse, error) {
	var document models.PatientDocument
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&document).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("documento não encontrado")
		}
		return nil, err
	}

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
func (s *DocumentService) Delete(associationID, id uuid.UUID) error {
	var document models.PatientDocument
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&document).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("documento não encontrado")
		}
		return err
	}

	// Remover do S3
	if _, err := s.s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(document.FileURL),
	}); err != nil {
		fmt.Printf("⚠️ erro ao remover arquivo do S3 (key=%s): %v\n", document.FileURL, err)
	}

	if err := s.db.Delete(&document).Error; err != nil {
		return err
	}

	return nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

func toDocumentResponse(document *models.PatientDocument) *DocumentResponse {
	return &DocumentResponse{
		ID:           document.ID.String(),
		PatientID:    document.PatientID.String(),
		DocumentType: document.DocumentType,
		FileName:     document.FileName,
		FileSize:     document.FileSize,
		MimeType:     document.MimeType,
		Status:       document.Status,
		CreatedAt:    document.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    document.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

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

func (s *DocumentService) isAllowedMimeType(mimeType string) bool {
	for _, t := range s.allowedTypes {
		if t == mimeType {
			return true
		}
	}
	return false
}

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