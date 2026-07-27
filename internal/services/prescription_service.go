// ================================================================
// PACOTE SERVICES - PRESCRIPTION SERVICE
// ================================================================
// Camada de serviço responsável pela gestão de prescrições/receitas.
//
// RESPONSABILIDADES:
// 1. CRUD completo de prescrições
// 2. Validação de receitas (datas, médico, paciente, produtos)
// 3. Alertas de vencimento (15 dias antes)
// 4. Controle de status (valida, proxima_vencer, vencida)
// 5. Validação para novos pedidos (receita válida)
// 6. Itens da receita (medicamentos e dosagens)
// ================================================================

package services

import (
	"errors"
	"fmt"
	"time"

	"cannacare-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT PRESCRIPTIONSERVICE
// ================================================================
type PrescriptionService struct {
	db *gorm.DB
}

// ================================================================
// FUNÇÃO NEWPRESCRIPTIONSERVICE()
// ================================================================
func NewPrescriptionService(db *gorm.DB) *PrescriptionService {
	return &PrescriptionService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

// CreatePrescriptionRequest - Dados para criar uma nova prescrição
type CreatePrescriptionRequest struct {
	PatientID      string                          `json:"patient_id" validate:"required"`
	DoctorID       string                          `json:"doctor_id" validate:"required"`
	CID            string                          `json:"cid" validate:"required"`
	IssueDate      time.Time                       `json:"issue_date" validate:"required"`
	ExpirationDate time.Time                       `json:"expiration_date" validate:"required"`
	Items          []CreatePrescriptionItemRequest `json:"items" validate:"required,min=1"`
}

// CreatePrescriptionItemRequest - Item da prescrição
type CreatePrescriptionItemRequest struct {
	ProductID           string `json:"product_id" validate:"required"`
	DosageInstructions  string `json:"dosage_instructions" validate:"required"`
	QuantityRecommended int    `json:"quantity_recommended" validate:"required,min=1"`
}

// UpdatePrescriptionRequest - Dados para atualizar uma prescrição
type UpdatePrescriptionRequest struct {
	CID            string    `json:"cid" validate:"omitempty"`
	IssueDate      time.Time `json:"issue_date" validate:"omitempty"`
	ExpirationDate time.Time `json:"expiration_date" validate:"omitempty"`
	IsActive       *bool     `json:"is_active" validate:"omitempty"`
}

// PrescriptionResponse - Resposta com dados da prescrição
type PrescriptionResponse struct {
	ID              string                     `json:"id"`
	PatientID       string                     `json:"patient_id"`
	PatientName     string                     `json:"patient_name"`
	DoctorID        string                     `json:"doctor_id"`
	DoctorName      string                     `json:"doctor_name"`
	CID             string                     `json:"cid"`
	IssueDate       string                     `json:"issue_date"`
	ExpirationDate  string                     `json:"expiration_date"`
	Status          string                     `json:"status"`
	IsActive        bool                       `json:"is_active"`
	Items           []PrescriptionItemResponse `json:"items"`
	DaysUntilExpire int                        `json:"days_until_expire"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
}

// PrescriptionItemResponse - Item da prescrição
type PrescriptionItemResponse struct {
	ID                  string `json:"id"`
	ProductID           string `json:"product_id"`
	ProductName         string `json:"product_name"`
	DosageInstructions  string `json:"dosage_instructions"`
	QuantityRecommended int    `json:"quantity_recommended"`
}

// ListPrescriptionRequest - Filtros para listagem
type ListPrescriptionRequest struct {
	PatientID string `json:"patient_id" query:"patient_id"`
	DoctorID  string `json:"doctor_id" query:"doctor_id"`
	Status    string `json:"status" query:"status"`
	IsActive  *bool  `json:"is_active" query:"is_active"`
	Page      int    `json:"page" query:"page"`
	Limit     int    `json:"limit" query:"limit"`
}

// ValidatePrescriptionResult - Resultado da validação de receita
type ValidatePrescriptionResult struct {
	IsValid        bool   `json:"is_valid"`
	PrescriptionID string `json:"prescription_id,omitempty"`
	Message        string `json:"message"`
}

// ================================================================
// FUNÇÃO CREATE()
// ================================================================
// Cria uma nova prescrição com seus itens
func (s *PrescriptionService) Create(req CreatePrescriptionRequest) (*PrescriptionResponse, error) {
	// --- 1. Validar datas ---
	if req.ExpirationDate.Before(req.IssueDate) {
		return nil, errors.New("data de validade não pode ser anterior à data de emissão")
	}
	if req.IssueDate.After(time.Now()) {
		return nil, errors.New("data de emissão não pode ser futura")
	}

	// --- 2. Validar paciente ---
	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		return nil, fmt.Errorf("ID do paciente inválido: %w", err)
	}
	var patient models.Patient
	if err := s.db.Where("id = ?", patientID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}
	if patient.Status != "aprovado" {
		return nil, errors.New("paciente não está aprovado")
	}

	// --- 3. Validar médico ---
	doctorID, err := uuid.Parse(req.DoctorID)
	if err != nil {
		return nil, fmt.Errorf("ID do médico inválido: %w", err)
	}
	var doctor models.Doctor
	if err := s.db.Where("id = ? AND is_active = ?", doctorID, true).First(&doctor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("médico não encontrado ou inativo")
		}
		return nil, err
	}

	// --- 4. Validar produtos e criar prescrição ---
	var items []models.PrescriptionItem
	var productIDs []uuid.UUID

	for _, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("ID do produto inválido: %w", err)
		}
		productIDs = append(productIDs, productID)

		items = append(items, models.PrescriptionItem{
			ProductID:           productID,
			DosageInstructions:  item.DosageInstructions,
			QuantityRecommended: item.QuantityRecommended,
		})
	}

	// Verificar se todos os produtos existem
	var products []models.Product
	if err := s.db.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return nil, err
	}
	if len(products) != len(productIDs) {
		return nil, errors.New("um ou mais produtos não encontrados")
	}

	// --- 5. Definir status da prescrição ---
	status := "valida"
	if req.ExpirationDate.Before(time.Now()) {
		status = "vencida"
	} else if req.ExpirationDate.Before(time.Now().AddDate(0, 0, 15)) {
		status = "proxima_vencer"
	}

	    // --- 6. Criar a prescrição ---
    prescription := &models.Prescription{
        PatientID:      patientID,
        DoctorID:       doctorID,
        CID:            req.CID,
        IssueDate:      req.IssueDate,
        ExpirationDate: req.ExpirationDate,
        Status:         status,
        IsActive:       true,
        Items:          items,
    }

    if err := s.db.Create(prescription).Error; err != nil {
        return nil, err
    }

    // ============================================================
    // 🔧 CORREÇÃO: Recarregar a prescrição com os relacionamentos
    // ============================================================
    var createdPrescription models.Prescription
    if err := s.db.Preload("Items").Preload("Items.Product").Preload("Patient").Preload("Doctor").
        Where("id = ?", prescription.ID).First(&createdPrescription).Error; err != nil {
        return nil, err
    }

    // --- 7. Retornar resposta ---
    return s.toPrescriptionResponse(&createdPrescription), nil
}

// ================================================================
// FUNÇÃO GETBYID()
// ================================================================
// Busca uma prescrição pelo ID
func (s *PrescriptionService) GetByID(id uuid.UUID) (*PrescriptionResponse, error) {
	var prescription models.Prescription
	if err := s.db.Preload("Items").Preload("Items.Product").Preload("Patient").Preload("Doctor").
		Where("id = ?", id).First(&prescription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("prescrição não encontrada")
		}
		return nil, err
	}
	return s.toPrescriptionResponse(&prescription), nil
}

// ================================================================
// FUNÇÃO LIST()
// ================================================================
// Lista prescrições com filtros e paginação
func (s *PrescriptionService) List(req ListPrescriptionRequest) ([]PrescriptionResponse, int64, error) {
	// --- 1. Definir paginação ---
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	// --- 2. Construir query ---
	query := s.db.Model(&models.Prescription{})

	if req.PatientID != "" {
		patientID, err := uuid.Parse(req.PatientID)
		if err != nil {
			return nil, 0, fmt.Errorf("ID do paciente inválido: %w", err)
		}
		query = query.Where("patient_id = ?", patientID)
	}
	if req.DoctorID != "" {
		doctorID, err := uuid.Parse(req.DoctorID)
		if err != nil {
			return nil, 0, fmt.Errorf("ID do médico inválido: %w", err)
		}
		query = query.Where("doctor_id = ?", doctorID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// --- 3. Contar total ---
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// --- 4. Buscar com paginação ---
	var prescriptions []models.Prescription
	if err := query.Preload("Items").Preload("Items.Product").Preload("Patient").Preload("Doctor").
		Offset(offset).Limit(req.Limit).Order("created_at DESC").Find(&prescriptions).Error; err != nil {
		return nil, 0, err
	}

	// --- 5. Converter para resposta ---
	var responses []PrescriptionResponse
	for _, p := range prescriptions {
		responses = append(responses, *s.toPrescriptionResponse(&p))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATE()
// ================================================================
// Atualiza os dados de uma prescrição
func (s *PrescriptionService) Update(id uuid.UUID, req UpdatePrescriptionRequest) (*PrescriptionResponse, error) {
	var prescription models.Prescription
	if err := s.db.Where("id = ?", id).First(&prescription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("prescrição não encontrada")
		}
		return nil, err
	}

	// --- 1. Atualizar campos ---
	if req.CID != "" {
		prescription.CID = req.CID
	}
	if !req.IssueDate.IsZero() {
		if req.IssueDate.After(time.Now()) {
			return nil, errors.New("data de emissão não pode ser futura")
		}
		prescription.IssueDate = req.IssueDate
	}
	if !req.ExpirationDate.IsZero() {
		if req.ExpirationDate.Before(prescription.IssueDate) {
			return nil, errors.New("data de validade não pode ser anterior à data de emissão")
		}
		prescription.ExpirationDate = req.ExpirationDate
	}
	if req.IsActive != nil {
		prescription.IsActive = *req.IsActive
	}

	// --- 2. Atualizar status automaticamente ---
	if req.ExpirationDate.IsZero() {
		s.updateStatus(&prescription)
	} else {
		if prescription.ExpirationDate.Before(time.Now()) {
			prescription.Status = "vencida"
		} else if prescription.ExpirationDate.Before(time.Now().AddDate(0, 0, 15)) {
			prescription.Status = "proxima_vencer"
		} else {
			prescription.Status = "valida"
		}
	}

	if err := s.db.Save(&prescription).Error; err != nil {
		return nil, err
	}

	// --- 3. Carregar relacionamentos ---
	if err := s.db.Preload("Items").Preload("Items.Product").Preload("Patient").Preload("Doctor").
		First(&prescription, id).Error; err != nil {
		return nil, err
	}

	return s.toPrescriptionResponse(&prescription), nil
}

// ================================================================
// FUNÇÃO DELETE()
// ================================================================
// Remove uma prescrição (soft delete)
func (s *PrescriptionService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.Prescription{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("prescrição não encontrada")
	}
	return nil
}

// ================================================================
// FUNÇÃO VALIDATE()
// ================================================================
// Valida se uma prescrição é válida para um novo pedido
// Verifica: existe, ativa, não vencida, paciente aprovado
func (s *PrescriptionService) Validate(prescriptionID uuid.UUID) (*ValidatePrescriptionResult, error) {
	var prescription models.Prescription
	if err := s.db.Where("id = ?", prescriptionID).First(&prescription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ValidatePrescriptionResult{
				IsValid: false,
				Message: "prescrição não encontrada",
			}, nil
		}
		return nil, err
	}

	// Verificar se está ativa
	if !prescription.IsActive {
		return &ValidatePrescriptionResult{
			IsValid: false,
			Message: "prescrição inativa",
		}, nil
	}

	// Verificar se está vencida
	if prescription.Status == "vencida" {
		return &ValidatePrescriptionResult{
			IsValid: false,
			Message: "prescrição vencida em " + prescription.ExpirationDate.Format("02/01/2006"),
		}, nil
	}

	// Verificar se o paciente está ativo
	var patient models.Patient
	if err := s.db.Where("id = ?", prescription.PatientID).First(&patient).Error; err != nil {
		return nil, err
	}
	if patient.Status != "aprovado" {
		return &ValidatePrescriptionResult{
			IsValid: false,
			Message: "paciente não está aprovado",
		}, nil
	}

	// Verificar se o médico está ativo
	var doctor models.Doctor
	if err := s.db.Where("id = ? AND is_active = ?", prescription.DoctorID, true).First(&doctor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ValidatePrescriptionResult{
				IsValid: false,
				Message: "médico não está ativo",
			}, nil
		}
		return nil, err
	}

	return &ValidatePrescriptionResult{
		IsValid:        true,
		PrescriptionID: prescription.ID.String(),
		Message:        "prescrição válida",
	}, nil
}

// ================================================================
// FUNÇÃO GETEXPIRED()
// ================================================================
// Retorna prescrições vencidas (usando view)
func (s *PrescriptionService) GetExpired() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_expired_prescriptions").Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ================================================================
// FUNÇÃO UPDATEALLSTATUS()
// ================================================================
// Atualiza o status de todas as prescrições (job diário)
func (s *PrescriptionService) UpdateAllStatus() error {
	var prescriptions []models.Prescription
	if err := s.db.Where("is_active = ?", true).Find(&prescriptions).Error; err != nil {
		return err
	}

	for _, p := range prescriptions {
		s.updateStatus(&p)
		if err := s.db.Save(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// updateStatus - Atualiza o status da prescrição baseado na data de validade
func (s *PrescriptionService) updateStatus(p *models.Prescription) {
	now := time.Now()
	if p.ExpirationDate.Before(now) {
		p.Status = "vencida"
		p.IsActive = false
	} else if p.ExpirationDate.Before(now.AddDate(0, 0, 15)) {
		p.Status = "proxima_vencer"
	} else {
		p.Status = "valida"
	}
}

// toPrescriptionResponse - Converte models.Prescription para PrescriptionResponse
// FUNÇÃO toPrescriptionResponse()
func (s *PrescriptionService) toPrescriptionResponse(p *models.Prescription) *PrescriptionResponse {
    // Calcular dias até vencer
    daysUntilExpire := int(p.ExpirationDate.Sub(time.Now()).Hours() / 24)
    if daysUntilExpire < 0 {
        daysUntilExpire = 0
    }

    items := []PrescriptionItemResponse{}
    for _, item := range p.Items {
        productName := ""
        // 🔧 Verificar se Product não é nil
        if item.Product.ID != uuid.Nil {
            productName = item.Product.Name
        }
        items = append(items, PrescriptionItemResponse{
            ID:                  item.ID.String(),
            ProductID:           item.ProductID.String(),
            ProductName:         productName,
            DosageInstructions:  item.DosageInstructions,
            QuantityRecommended: item.QuantityRecommended,
        })
    }

    patientName := ""
    // 🔧 Verificar se Patient não é nil
    if p.Patient.ID != uuid.Nil {
        patientName = p.Patient.FullName
    }

    doctorName := ""
    // 🔧 Verificar se Doctor não é nil
    if p.Doctor.ID != uuid.Nil {
        doctorName = p.Doctor.Name
    }

    return &PrescriptionResponse{
        ID:              p.ID.String(),
        PatientID:       p.PatientID.String(),
        PatientName:     patientName,
        DoctorID:        p.DoctorID.String(),
        DoctorName:      doctorName,
        CID:             p.CID,
        IssueDate:       p.IssueDate.Format("2006-01-02"),
        ExpirationDate:  p.ExpirationDate.Format("2006-01-02"),
        Status:          p.Status,
        IsActive:        p.IsActive,
        Items:           items,
        DaysUntilExpire: daysUntilExpire,
        CreatedAt:       p.CreatedAt.Format("2006-01-02 15:04:05"),
        UpdatedAt:       p.UpdatedAt.Format("2006-01-02 15:04:05"),
    }
}
