// ================================================================
// PACOTE SERVICES - DOCTOR SERVICE (COM MULTI-TENANCY)
// ================================================================
// Camada de serviço responsável pela gestão de médicos.
// ================================================================

package services

import (
	"cannacare-backend/internal/models"
	"cannacare-backend/internal/utils"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT DOCTORSERVICE
// ================================================================
type DoctorService struct {
	db *gorm.DB
}

// ================================================================
// FUNÇÃO NEWDOCTORSERVICE()
// ================================================================
func NewDoctorService(db *gorm.DB) *DoctorService {
	return &DoctorService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

type CreateDoctorRequest struct {
	Name      string `json:"name" validate:"required,min=3,max=100"`
	CRM       string `json:"crm" validate:"required,min=4,max=20"`
	CRMState  string `json:"crm_state" validate:"required,min=2,max=2"`
	Specialty string `json:"specialty" validate:"omitempty"`
	Phone     string `json:"phone" validate:"omitempty"`
	Email     string `json:"email" validate:"omitempty,email"`
}

type UpdateDoctorRequest struct {
	Name      string `json:"name" validate:"omitempty,min=3,max=100"`
	Specialty string `json:"specialty" validate:"omitempty"`
	Phone     string `json:"phone" validate:"omitempty"`
	Email     string `json:"email" validate:"omitempty,email"`
	IsActive  *bool  `json:"is_active" validate:"omitempty"`
}

type DoctorResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CRM        string `json:"crm"`
	CRMState   string `json:"crm_state"`
	Specialty  string `json:"specialty,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Email      string `json:"email,omitempty"`
	IsActive   bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type ListDoctorRequest struct {
	Name      string `json:"name" query:"name"`
	CRM       string `json:"crm" query:"crm"`
	Specialty string `json:"specialty" query:"specialty"`
	IsActive  *bool  `json:"is_active" query:"is_active"`
	Page      int    `json:"page" query:"page"`
	Limit     int    `json:"limit" query:"limit"`
}

// ================================================================
// FUNÇÃO CREATE() - COM MULTI-TENANCY
// ================================================================
// ⚠️ AGORA RECEBE association_id COMO PRIMEIRO PARÂMETRO!
// ================================================================
func (s *DoctorService) Create(associationID uuid.UUID, req CreateDoctorRequest) (*DoctorResponse, error) {
	// --- 1. Validar dados ---
	if err := validateDoctorData(req); err != nil {
		return nil, err
	}

	// --- 2. Verificar se CRM já existe DENTRO da associação ---
	var existingDoctor models.Doctor
	err := s.db.Where("crm = ? AND crm_state = ? AND association_id = ?", 
		req.CRM, strings.ToUpper(req.CRMState), associationID).First(&existingDoctor).Error
	if err == nil {
		return nil, fmt.Errorf("médico com CRM %s-%s já cadastrado nesta associação", req.CRM, req.CRMState)
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- 3. Verificar se email já existe DENTRO da associação ---
	if req.Email != "" {
		err = s.db.Where("email = ? AND association_id = ?", req.Email, associationID).First(&existingDoctor).Error
		if err == nil {
			return nil, fmt.Errorf("email %s já cadastrado nesta associação", req.Email)
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	// --- 4. Criar o médico com association_id ---
	doctor := &models.Doctor{
		AssociationID: associationID, // ← ESSENCIAL!
		Name:          req.Name,
		CRM:           req.CRM,
		CRMState:      strings.ToUpper(req.CRMState),
		Specialty:     req.Specialty,
		Phone:         req.Phone,
		Email:         req.Email,
		IsActive:      true,
	}

	if err := s.db.Create(doctor).Error; err != nil {
		return nil, err
	}

	return toDoctorResponse(doctor), nil
}

// ================================================================
// FUNÇÃO GETBYID() - COM MULTI-TENANCY
// ================================================================
// ⚠️ AGORA RECEBE association_id COMO PRIMEIRO PARÂMETRO!
// ================================================================
func (s *DoctorService) GetByID(associationID uuid.UUID, id uuid.UUID) (*DoctorResponse, error) {
	var doctor models.Doctor
	// ⚠️ SEMPRE filtra por association_id!
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&doctor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("médico não encontrado")
		}
		return nil, err
	}
	return toDoctorResponse(&doctor), nil
}

// ================================================================
// FUNÇÃO LIST() - COM MULTI-TENANCY
// ================================================================
// ⚠️ AGORA RECEBE association_id COMO PRIMEIRO PARÂMETRO!
// ================================================================
func (s *DoctorService) List(associationID uuid.UUID, req ListDoctorRequest) ([]DoctorResponse, int64, error) {
	// --- 1. Definir valores padrão para paginação ---
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	// --- 2. Construir a query SEMPRE com association_id ---
	// ⚠️ NUNCA remova o filtro de association_id!
	query := s.db.Model(&models.Doctor{}).Where("association_id = ?", associationID)

	// Aplicar filtros (todos COM association_id)
	if req.Name != "" {
		query = query.Where("name ILIKE ?", "%"+req.Name+"%")
	}
	if req.CRM != "" {
		query = query.Where("crm ILIKE ?", "%"+req.CRM+"%")
	}
	if req.Specialty != "" {
		query = query.Where("specialty ILIKE ?", "%"+req.Specialty+"%")
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// --- 3. Contar total de registros ---
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// --- 4. Buscar registros com paginação ---
	var doctors []models.Doctor
	if err := query.Offset(offset).Limit(req.Limit).Order("name ASC").Find(&doctors).Error; err != nil {
		return nil, 0, err
	}

	// --- 5. Converter para resposta ---
	var responses []DoctorResponse
	for _, doctor := range doctors {
		responses = append(responses, *toDoctorResponse(&doctor))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATE() - COM MULTI-TENANCY
// ================================================================
// ⚠️ AGORA RECEBE association_id COMO PRIMEIRO PARÂMETRO!
// ================================================================
func (s *DoctorService) Update(associationID uuid.UUID, id uuid.UUID, req UpdateDoctorRequest) (*DoctorResponse, error) {
	// --- 1. Buscar o médico (SEMPRE com association_id) ---
	var doctor models.Doctor
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&doctor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("médico não encontrado")
		}
		return nil, err
	}

	// --- 2. Atualizar campos ---
	if req.Name != "" {
		doctor.Name = req.Name
	}
	if req.Specialty != "" {
		doctor.Specialty = req.Specialty
	}
	if req.Phone != "" {
		doctor.Phone = req.Phone
	}
	if req.Email != "" {
		if !utils.IsValidEmail(req.Email) {
			return nil, errors.New("email inválido")
		}
		doctor.Email = req.Email
	}
	if req.IsActive != nil {
		doctor.IsActive = *req.IsActive
	}

	// --- 3. Salvar no banco ---
	if err := s.db.Save(&doctor).Error; err != nil {
		return nil, err
	}

	return toDoctorResponse(&doctor), nil
}

// ================================================================
// FUNÇÃO DELETE() - COM MULTI-TENANCY
// ================================================================
// ⚠️ AGORA RECEBE association_id COMO PRIMEIRO PARÂMETRO!
// ================================================================
func (s *DoctorService) Delete(associationID uuid.UUID, id uuid.UUID) error {
	// ⚠️ SEMPRE filtra por association_id!
	result := s.db.Where("id = ? AND association_id = ?", id, associationID).Delete(&models.Doctor{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("médico não encontrado")
	}
	return nil
}

// ================================================================
// FUNÇÃO GETTOPDOCTORS() - COM MULTI-TENANCY
// ================================================================
// ⚠️ AGORA RECEBE association_id COMO PRIMEIRO PARÂMETRO!
// ================================================================
func (s *DoctorService) GetTopDoctors(associationID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// ⚠️ A view vw_top_doctors já tem association_id
	err := s.db.Table("vw_top_doctors").Where("association_id = ?", associationID).Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// toDoctorResponse - Converte models.Doctor para DoctorResponse
func toDoctorResponse(doctor *models.Doctor) *DoctorResponse {
	return &DoctorResponse{
		ID:         doctor.ID.String(),
		Name:       doctor.Name,
		CRM:        doctor.CRM,
		CRMState:   doctor.CRMState,
		Specialty:  doctor.Specialty,
		Phone:      doctor.Phone,
		Email:      doctor.Email,
		IsActive:   doctor.IsActive,
		CreatedAt:  doctor.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  doctor.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// validateDoctorData - Valida os dados do médico
func validateDoctorData(req CreateDoctorRequest) error {
	// Validar nome
	if len(req.Name) < 3 {
		return errors.New("nome deve ter pelo menos 3 caracteres")
	}
	if len(req.Name) > 100 {
		return errors.New("nome deve ter no máximo 100 caracteres")
	}

	// Validar CRM
	if len(req.CRM) < 4 {
		return errors.New("CRM inválido")
	}
	if len(req.CRM) > 20 {
		return errors.New("CRM deve ter no máximo 20 caracteres")
	}

	// Validar estado (2 letras)
	req.CRMState = strings.ToUpper(req.CRMState)
	if len(req.CRMState) != 2 {
		return errors.New("UF deve ter 2 caracteres (ex: SP, RJ, MG)")
	}

	// Validar email (se informado)
	if req.Email != "" && !utils.IsValidEmail(req.Email) {
		return errors.New("email inválido")
	}

	// Validar telefone (se informado) - opcional
	if req.Phone != "" && !utils.IsValidPhone(req.Phone) {
		return errors.New("telefone inválido")
	}

	return nil
}