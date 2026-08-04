// ================================================================
// CANNACARE - PATIENT SERVICE
// ================================================================
// Camada de serviço responsável pela lógica de negócio de pacientes.
//
// MULTI-TENANCY:
//   TODAS as operações recebem association_id como primeiro parâmetro.
//   Todas as queries SQL filtram por association_id.
//   NUNCA faça uma query sem filtro de associação!
// ================================================================

package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cannacare-backend/internal/models"
	"cannacare-backend/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT PATIENTSERVICE
// ================================================================
type PatientService struct {
	db *gorm.DB
}

// ================================================================
// FUNÇÃO NEWPATIENTSERVICE()
// ================================================================
func NewPatientService(db *gorm.DB) *PatientService {
	return &PatientService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

type CreatePatientRequest struct {
	FullName            string    `json:"full_name" validate:"required,min=3,max=200"`
	BirthDate           time.Time `json:"birth_date" validate:"required"`
	Gender              string    `json:"gender" validate:"omitempty,oneof=Masculino Feminino Outro"`
	CPF                 string    `json:"cpf" validate:"required,len=11"`
	RG                  string    `json:"rg" validate:"omitempty"`
	Phone               string    `json:"phone" validate:"omitempty"`
	WhatsApp            string    `json:"whatsapp" validate:"omitempty"`
	Email               string    `json:"email" validate:"omitempty,email"`
	AddressStreet       string    `json:"address_street" validate:"omitempty"`
	AddressNumber       string    `json:"address_number" validate:"omitempty"`
	AddressComplement   string    `json:"address_complement" validate:"omitempty"`
	AddressNeighborhood string    `json:"address_neighborhood" validate:"omitempty"`
	AddressCity         string    `json:"address_city" validate:"omitempty"`
	AddressState        string    `json:"address_state" validate:"omitempty,len=2"`
	AddressZipCode      string    `json:"address_zipcode" validate:"omitempty"`
	Status              string    `json:"status" validate:"omitempty,oneof=pendente_documentacao em_analise aprovado negado assistente_social"`
	IsSocialPatient     bool      `json:"is_social_patient"`
}

type UpdatePatientRequest struct {
	FullName            string    `json:"full_name" validate:"omitempty,min=3,max=200"`
	BirthDate           time.Time `json:"birth_date" validate:"omitempty"`
	Gender              string    `json:"gender" validate:"omitempty,oneof=Masculino Feminino Outro"`
	Phone               string    `json:"phone" validate:"omitempty"`
	WhatsApp            string    `json:"whatsapp" validate:"omitempty"`
	Email               string    `json:"email" validate:"omitempty,email"`
	AddressStreet       string    `json:"address_street" validate:"omitempty"`
	AddressNumber       string    `json:"address_number" validate:"omitempty"`
	AddressComplement   string    `json:"address_complement" validate:"omitempty"`
	AddressNeighborhood string    `json:"address_neighborhood" validate:"omitempty"`
	AddressCity         string    `json:"address_city" validate:"omitempty"`
	AddressState        string    `json:"address_state" validate:"omitempty,len=2"`
	AddressZipCode      string    `json:"address_zipcode" validate:"omitempty"`
	IsSocialPatient     *bool     `json:"is_social_patient" validate:"omitempty"`
}

// PatientResponse - Resposta com dados do paciente
type PatientResponse struct {
	ID                   string     `json:"id"`
	AssociationID        string     `json:"association_id"` // ← MULTI-TENANCY
	UserID               *string    `json:"user_id,omitempty"`
	FullName             string     `json:"full_name"`
	BirthDate            string     `json:"birth_date"`
	Gender               string     `json:"gender,omitempty"`
	CPF                  string     `json:"cpf"`
	RG                   string     `json:"rg,omitempty"`
	Phone                string     `json:"phone,omitempty"`
	WhatsApp             string     `json:"whatsapp,omitempty"`
	Email                string     `json:"email,omitempty"`
	AddressStreet        string     `json:"address_street,omitempty"`
	AddressNumber        string     `json:"address_number,omitempty"`
	AddressComplement    string     `json:"address_complement,omitempty"`
	AddressNeighborhood  string     `json:"address_neighborhood,omitempty"`
	AddressCity          string     `json:"address_city,omitempty"`
	AddressState         string     `json:"address_state,omitempty"`
	AddressZipCode       string     `json:"address_zipcode,omitempty"`
	Status               string     `json:"status"`
	IsSocialPatient      bool       `json:"is_social_patient"`
	SocialAssistantNotes string     `json:"social_assistant_notes,omitempty"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`
	CreatedAt            string     `json:"created_at"`
	UpdatedAt            string     `json:"updated_at"`
}

type ListPatientRequest struct {
	Name     string `json:"name" query:"name"`
	CPF      string `json:"cpf" query:"cpf"`
	Email    string `json:"email" query:"email"`
	Status   string `json:"status" query:"status"`
	IsSocial *bool  `json:"is_social" query:"is_social"`
	Page     int    `json:"page" query:"page"`
	Limit    int    `json:"limit" query:"limit"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pendente_documentacao em_analise aprovado negado assistente_social"`
	Reason string `json:"reason" validate:"omitempty"`
}

// ================================================================
// FUNÇÃO CREATE()
// ================================================================
// Cria um novo paciente na associação.
//
// PARÂMETROS:
//   - associationID: ID da associação (extraído do JWT)
//   - req: Dados do paciente
//
// RETORNO:
//   - *PatientResponse: Dados do paciente criado
//   - error: Erro se houver falha
//
// ⚠️ IMPORTANTE: O association_id é SEMPRE passado como parâmetro!
// Nunca confie no frontend para enviar o association_id.
// ================================================================
func (s *PatientService) Create(associationID uuid.UUID, req CreatePatientRequest) (*PatientResponse, error) {
	// --- PASSO 1: Validar dados ---
	if err := validatePatientData(req); err != nil {
		return nil, err
	}

	// --- PASSO 2: Verificar CPF duplicado DENTRO da associação ---
	var existingPatient models.Patient
	err := s.db.Where("cpf = ? AND association_id = ?", req.CPF, associationID).First(&existingPatient).Error
	if err == nil {
		return nil, fmt.Errorf("CPF %s já cadastrado nesta associação", req.CPF)
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- PASSO 3: Verificar email duplicado DENTRO da associação ---
	if req.Email != "" {
		err = s.db.Where("email = ? AND association_id = ?", req.Email, associationID).First(&existingPatient).Error
		if err == nil {
			return nil, fmt.Errorf("email %s já cadastrado nesta associação", req.Email)
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	// --- PASSO 4: Verificar se a associação atingiu o limite de pacientes ---
	var association models.Association
	if err := s.db.Where("id = ?", associationID).First(&association).Error; err != nil {
		return nil, err
	}

	var patientCount int64
	s.db.Model(&models.Patient{}).Where("association_id = ?", associationID).Count(&patientCount)

	if association.Plan == "basic" && int(patientCount) >= association.PatientLimit {
		return nil, errors.New("limite de pacientes do plano básico atingido. Faça upgrade para o plano premium")
	}

	// --- PASSO 5: Criar o paciente com association_id ---
	patient := &models.Patient{
		AssociationID:       associationID, // ← MULTI-TENANCY: sempre setado!
		FullName:            req.FullName,
		BirthDate:           req.BirthDate,
		Gender:              req.Gender,
		CPF:                 req.CPF,
		RG:                  req.RG,
		Phone:               req.Phone,
		WhatsApp:            req.WhatsApp,
		Email:               req.Email,
		AddressStreet:       req.AddressStreet,
		AddressNumber:       req.AddressNumber,
		AddressComplement:   req.AddressComplement,
		AddressNeighborhood: req.AddressNeighborhood,
		AddressCity:         req.AddressCity,
		AddressState:        strings.ToUpper(req.AddressState),
		AddressZipCode:      req.AddressZipCode,
		Status:              "pendente_documentacao",
		IsSocialPatient:     req.IsSocialPatient,
	}

	if err := s.db.Create(patient).Error; err != nil {
		return nil, err
	}

	return toPatientResponse(patient), nil
}

// ================================================================
// FUNÇÃO GETBYID()
// ================================================================
// Busca um paciente pelo ID (SEMPRE com filtro de associação)
//
// ⚠️ IMPORTANTE: NUNCA faça: db.Where("id = ?", id)
// Sempre filtre por association_id também!
// ================================================================
func (s *PatientService) GetByID(associationID uuid.UUID, id uuid.UUID) (*PatientResponse, error) {
	var patient models.Patient
	// ⚠️ SEMPRE filtra por association_id!
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}
	return toPatientResponse(&patient), nil
}

// ================================================================
// FUNÇÃO LIST()
// ================================================================
// Lista pacientes da associação com filtros.
//
// PARÂMETROS:
//   - associationID: ID da associação (extraído do JWT)
//   - req: Filtros e paginação
//
// ⚠️ IMPORTANTE: SEMPRE filtra por association_id!
// Nunca faça: s.db.Find(&patients) - Isso mostraria pacientes de TODAS as associações!
// ================================================================
func (s *PatientService) List(associationID uuid.UUID, req ListPatientRequest) ([]PatientResponse, int64, error) {
	// --- PASSO 1: Definir paginação ---
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	// --- PASSO 2: Construir query SEMPRE com association_id ---
	// ⚠️ NUNCA remova o filtro de association_id!
	query := s.db.Model(&models.Patient{}).Where("association_id = ?", associationID)

	// Aplicar filtros (todos COM association_id)
	if req.Name != "" {
		query = query.Where("full_name ILIKE ?", "%"+req.Name+"%")
	}
	if req.CPF != "" {
		query = query.Where("cpf ILIKE ?", "%"+req.CPF+"%")
	}
	if req.Email != "" {
		query = query.Where("email ILIKE ?", "%"+req.Email+"%")
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.IsSocial != nil {
		query = query.Where("is_social_patient = ?", *req.IsSocial)
	}

	// --- PASSO 3: Contar total (SEMPRE com association_id) ---
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// --- PASSO 4: Buscar com paginação (SEMPRE com association_id) ---
	var patients []models.Patient
	if err := query.Offset(offset).Limit(req.Limit).Order("full_name ASC").Find(&patients).Error; err != nil {
		return nil, 0, err
	}

	// --- PASSO 5: Converter para resposta ---
	var responses []PatientResponse
	for _, patient := range patients {
		responses = append(responses, *toPatientResponse(&patient))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATE()
// ================================================================
// Atualiza os dados de um paciente (SEMPRE com filtro de associação)
// ================================================================
func (s *PatientService) Update(associationID uuid.UUID, id uuid.UUID, req UpdatePatientRequest) (*PatientResponse, error) {
	// --- PASSO 1: Buscar paciente (SEMPRE com association_id) ---
	var patient models.Patient
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}

	// --- PASSO 2: Atualizar campos ---
	if req.FullName != "" {
		patient.FullName = req.FullName
	}
	if !req.BirthDate.IsZero() {
		patient.BirthDate = req.BirthDate
	}
	if req.Gender != "" {
		patient.Gender = req.Gender
	}
	if req.Phone != "" {
		patient.Phone = req.Phone
	}
	if req.WhatsApp != "" {
		patient.WhatsApp = req.WhatsApp
	}
	if req.Email != "" {
		if !utils.IsValidEmail(req.Email) {
			return nil, errors.New("email inválido")
		}
		patient.Email = req.Email
	}
	if req.AddressStreet != "" {
		patient.AddressStreet = req.AddressStreet
	}
	if req.AddressNumber != "" {
		patient.AddressNumber = req.AddressNumber
	}
	if req.AddressComplement != "" {
		patient.AddressComplement = req.AddressComplement
	}
	if req.AddressNeighborhood != "" {
		patient.AddressNeighborhood = req.AddressNeighborhood
	}
	if req.AddressCity != "" {
		patient.AddressCity = req.AddressCity
	}
	if req.AddressState != "" {
		patient.AddressState = strings.ToUpper(req.AddressState)
	}
	if req.AddressZipCode != "" {
		patient.AddressZipCode = req.AddressZipCode
	}
	if req.IsSocialPatient != nil {
		patient.IsSocialPatient = *req.IsSocialPatient
	}

	// --- PASSO 3: Salvar ---
	if err := s.db.Save(&patient).Error; err != nil {
		return nil, err
	}

	return toPatientResponse(&patient), nil
}

// ================================================================
// FUNÇÃO DELETE()
// ================================================================
// Remove um paciente (soft delete) - SEMPRE com filtro de associação
// ================================================================
func (s *PatientService) Delete(associationID uuid.UUID, id uuid.UUID) error {
	result := s.db.Where("id = ? AND association_id = ?", id, associationID).Delete(&models.Patient{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("paciente não encontrado")
	}
	return nil
}

// ================================================================
// FUNÇÃO UPDATESTATUS()
// ================================================================
// Atualiza o status de um paciente - SEMPRE com filtro de associação
// ================================================================
func (s *PatientService) UpdateStatus(associationID uuid.UUID, id uuid.UUID, req UpdateStatusRequest, userID uuid.UUID) (*PatientResponse, error) {
	// --- PASSO 1: Buscar paciente (SEMPRE com association_id) ---
	var patient models.Patient
	if err := s.db.Where("id = ? AND association_id = ?", id, associationID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}

	// --- PASSO 2: Verificar se o status é diferente do atual ---
	if patient.Status == req.Status {
		return nil, errors.New("paciente já está com este status")
	}

	// --- PASSO 3: Registrar histórico de status ---
	history := &models.PatientStatusHistory{
		AssociationID: associationID, // ← ESSENCIAL!
		PatientID:     patient.ID,
		ChangedBy:     &userID,
		OldStatus:     patient.Status,
		NewStatus:     req.Status,
		Reason:        req.Reason,
	}

	if err := s.db.Create(history).Error; err != nil {
		return nil, err
	}

	// --- PASSO 4: Atualizar status do paciente ---
	patient.Status = req.Status

	if req.Status == "aprovado" {
		now := time.Now()
		patient.ApprovedAt = &now
	}

	if err := s.db.Save(&patient).Error; err != nil {
		return nil, err
	}

	return toPatientResponse(&patient), nil
}

// ================================================================
// FUNÇÃO GETSTATISTICS()
// ================================================================
// Retorna estatísticas dos pacientes (dashboard)
// ================================================================
func (s *PatientService) GetStatistics(associationID uuid.UUID) (map[string]interface{}, error) {
	var stats map[string]interface{}
	// ⚠️ A view vw_patient_dashboard já tem association_id
	// Filtramos para trazer apenas os dados da associação
	err := s.db.Table("vw_patient_dashboard").Where("association_id = ?", associationID).Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// toPatientResponse - Converte models.Patient para PatientResponse
func toPatientResponse(patient *models.Patient) *PatientResponse {
	var userID *string
	if patient.UserID != nil {
		id := patient.UserID.String()
		userID = &id
	}

	return &PatientResponse{
		ID:                   patient.ID.String(),
		AssociationID:        patient.AssociationID.String(), // ← MULTI-TENANCY
		UserID:               userID,
		FullName:             patient.FullName,
		BirthDate:            patient.BirthDate.Format("2006-01-02"),
		Gender:               patient.Gender,
		CPF:                  patient.CPF,
		RG:                   patient.RG,
		Phone:                patient.Phone,
		WhatsApp:             patient.WhatsApp,
		Email:                patient.Email,
		AddressStreet:        patient.AddressStreet,
		AddressNumber:        patient.AddressNumber,
		AddressComplement:    patient.AddressComplement,
		AddressNeighborhood:  patient.AddressNeighborhood,
		AddressCity:          patient.AddressCity,
		AddressState:         patient.AddressState,
		AddressZipCode:       patient.AddressZipCode,
		Status:               patient.Status,
		IsSocialPatient:      patient.IsSocialPatient,
		SocialAssistantNotes: patient.SocialAssistantNotes,
		ApprovedAt:           patient.ApprovedAt,
		CreatedAt:            patient.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:            patient.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// validatePatientData - Valida os dados do paciente
func validatePatientData(req CreatePatientRequest) error {
	if len(req.FullName) < 3 {
		return errors.New("nome deve ter pelo menos 3 caracteres")
	}
	if len(req.FullName) > 200 {
		return errors.New("nome deve ter no máximo 200 caracteres")
	}

	req.CPF = strings.ReplaceAll(req.CPF, ".", "")
	req.CPF = strings.ReplaceAll(req.CPF, "-", "")
	if len(req.CPF) != 11 {
		return errors.New("CPF deve ter 11 dígitos")
	}
	if !utils.IsValidCPF(req.CPF) {
		return errors.New("CPF inválido")
	}

	if req.BirthDate.After(time.Now()) {
		return errors.New("data de nascimento não pode ser futura")
	}
	age := time.Now().Year() - req.BirthDate.Year()
	if age < 18 {
		return errors.New("paciente deve ter pelo menos 18 anos")
	}
	if age > 120 {
		return errors.New("data de nascimento inválida")
	}

	if req.Email != "" && !utils.IsValidEmail(req.Email) {
		return errors.New("email inválido")
	}

	if req.Phone != "" && !utils.IsValidPhone(req.Phone) {
		return errors.New("telefone inválido")
	}

	if req.WhatsApp != "" && !utils.IsValidPhone(req.WhatsApp) {
		return errors.New("whatsapp inválido")
	}

	if req.AddressState != "" {
		if !utils.IsValidState(req.AddressState) {
			return errors.New("UF inválida")
		}
	}

	return nil
}