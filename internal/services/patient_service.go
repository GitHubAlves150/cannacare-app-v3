// ================================================================
// PACOTE SERVICES - PATIENT SERVICE
// ================================================================
// Camada de serviço responsável pela gestão de pacientes.
//
// RESPONSABILIDADES:
// 1. CRUD completo de pacientes
// 2. Validação de dados (CPF, email, telefone, data de nascimento)
// 3. Busca com filtros (nome, CPF, status, etc)
// 4. Controle de status do paciente
// 5. Associação com usuário (User) para acesso ao portal
// 6. Paciente social (isenção de anuidade)
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
	"golang.org/x/crypto/bcrypt"
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
// Cria uma nova instância do serviço de pacientes
func NewPatientService(db *gorm.DB) *PatientService {
	return &PatientService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

// CreatePatientRequest - Dados para criar um novo paciente
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

// UpdatePatientRequest - Dados para atualizar um paciente
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

// PatientResponse - Resposta com dados do paciente (formato padronizado)
type PatientResponse struct {
	ID                   string     `json:"id"`
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

// ListPatientRequest - Filtros para listagem de pacientes
type ListPatientRequest struct {
	Name     string `json:"name" query:"name"`
	CPF      string `json:"cpf" query:"cpf"`
	Email    string `json:"email" query:"email"`
	Status   string `json:"status" query:"status"`
	IsSocial *bool  `json:"is_social" query:"is_social"`
	Page     int    `json:"page" query:"page"`
	Limit    int    `json:"limit" query:"limit"`
}

// UpdateStatusRequest - Para mudar status do paciente
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pendente_documentacao em_analise aprovado negado assistente_social"`
	Reason string `json:"reason" validate:"omitempty"`
}

// ================================================================
// FUNÇÃO CREATE()
// ================================================================
// Cria um novo paciente no sistema
//
// VALIDAÇÕES:
// 1. CPF único
// 2. Email único (se informado)
// 3. Data de nascimento válida (maior de 18 anos)
// 4. Telefone válido
// 5. UF válida
//
// FLUXO:
// 1. Validar dados
// 2. Verificar CPF duplicado
// 3. Verificar email duplicado
// 4. Criar usuário (User) para acesso ao portal
// 5. Criar paciente com vínculo ao usuário
// 6. Retornar resposta
func (s *PatientService) Create(req CreatePatientRequest) (*PatientResponse, error) {
	// --- 1. Validar dados ---
	if err := validatePatientData(req); err != nil {
		return nil, err
	}

	// --- 2. Verificar CPF duplicado ---
	var existingPatient models.Patient
	err := s.db.Where("cpf = ?", req.CPF).First(&existingPatient).Error
	if err == nil {
		return nil, fmt.Errorf("CPF %s já cadastrado", req.CPF)
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// --- 3. Verificar email duplicado (se informado) ---
	if req.Email != "" {
		err = s.db.Where("email = ?", req.Email).First(&existingPatient).Error
		if err == nil {
			return nil, fmt.Errorf("email %s já cadastrado para outro paciente", req.Email)
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	// --- 4. Criar usuário para acesso ao portal ---
	userID, err := s.createUserForPatient(req.FullName, req.Email)
	if err != nil {
		return nil, err
	}

	// --- 5. Definir status padrão ---
	status := req.Status
	if status == "" {
		status = "pendente_documentacao"
	}

	// --- 6. Criar o paciente ---
	patient := &models.Patient{
		UserID:              userID,
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
		Status:              status,
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
// Busca um paciente pelo ID
func (s *PatientService) GetByID(id uuid.UUID) (*PatientResponse, error) {
	var patient models.Patient
	if err := s.db.Where("id = ?", id).First(&patient).Error; err != nil {
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
// Lista pacientes com filtros e paginação
func (s *PatientService) List(req ListPatientRequest) ([]PatientResponse, int64, error) {
	// --- 1. Definir valores padrão para paginação ---
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	// --- 2. Construir a query ---
	query := s.db.Model(&models.Patient{})

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

	// --- 3. Contar total ---
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// --- 4. Buscar com paginação ---
	var patients []models.Patient
	if err := query.Offset(offset).Limit(req.Limit).Order("full_name ASC").Find(&patients).Error; err != nil {
		return nil, 0, err
	}

	// --- 5. Converter para resposta ---
	var responses []PatientResponse
	for _, patient := range patients {
		responses = append(responses, *toPatientResponse(&patient))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATE()
// ================================================================
// Atualiza os dados de um paciente
func (s *PatientService) Update(id uuid.UUID, req UpdatePatientRequest) (*PatientResponse, error) {
	// --- 1. Buscar paciente ---
	var patient models.Patient
	if err := s.db.Where("id = ?", id).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}

	// --- 2. Atualizar campos ---
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

	// --- 3. Salvar ---
	if err := s.db.Save(&patient).Error; err != nil {
		return nil, err
	}

	return toPatientResponse(&patient), nil
}

// ================================================================
// FUNÇÃO DELETE()
// ================================================================
// Remove um paciente (soft delete)
func (s *PatientService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.Patient{}, "id = ?", id)
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
// Atualiza o status de um paciente e registra no histórico
func (s *PatientService) UpdateStatus(id uuid.UUID, req UpdateStatusRequest, userID uuid.UUID) (*PatientResponse, error) {
	// --- 1. Buscar paciente ---
	var patient models.Patient
	if err := s.db.Where("id = ?", id).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}

	// --- 2. Verificar se o status é diferente do atual ---
	if patient.Status == req.Status {
		return nil, errors.New("paciente já está com este status")
	}

	// --- 3. Registrar histórico de status ---
	history := &models.PatientStatusHistory{
		PatientID: patient.ID,
		ChangedBy: &userID,
		OldStatus: patient.Status,
		NewStatus: req.Status,
		Reason:    req.Reason,
	}

	if err := s.db.Create(history).Error; err != nil {
		return nil, err
	}

	// --- 4. Atualizar status do paciente ---
	patient.Status = req.Status

	// Se for aprovado, registrar data de aprovação
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
func (s *PatientService) GetStatistics() (map[string]interface{}, error) {
	var stats map[string]interface{}

	err := s.db.Table("vw_patient_dashboard").Find(&stats).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// createUserForPatient - Cria um usuário para o paciente acessar o portal
func (s *PatientService) createUserForPatient(name, email string) (*uuid.UUID, error) {
	// Se não tiver email, não cria usuário
	if email == "" {
		return nil, nil
	}

	// Verificar se o email já existe
	var existingUser models.User
	err := s.db.Where("email = ?", email).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("email %s já está em uso", email)
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Gerar senha aleatória temporária (será trocada pelo paciente)
	tempPassword := generateTempPassword()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Criar usuário
	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "paciente",
		IsActive:     true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return &user.ID, nil
}

// generateTempPassword - Gera uma senha temporária aleatória
func generateTempPassword() string {
	// Senha temporária simples (será trocada pelo paciente)
	return "temp123456"
}

// toPatientResponse - Converte models.Patient para PatientResponse
func toPatientResponse(patient *models.Patient) *PatientResponse {
	var userID *string
	if patient.UserID != nil {
		id := patient.UserID.String()
		userID = &id
	}

	return &PatientResponse{
		ID:                   patient.ID.String(),
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
	// Validar nome
	if len(req.FullName) < 3 {
		return errors.New("nome deve ter pelo menos 3 caracteres")
	}
	if len(req.FullName) > 200 {
		return errors.New("nome deve ter no máximo 200 caracteres")
	}

	// Validar CPF (11 dígitos)
	req.CPF = strings.ReplaceAll(req.CPF, ".", "")
	req.CPF = strings.ReplaceAll(req.CPF, "-", "")
	if len(req.CPF) != 11 {
		return errors.New("CPF deve ter 11 dígitos")
	}
	if !utils.IsValidCPF(req.CPF) {
		return errors.New("CPF inválido")
	}

	// Validar data de nascimento (maior de 18 anos)
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

	// Validar email (se informado)
	if req.Email != "" && !utils.IsValidEmail(req.Email) {
		return errors.New("email inválido")
	}

	// Validar telefone (se informado)
	if req.Phone != "" && !utils.IsValidPhone(req.Phone) {
		return errors.New("telefone inválido")
	}

	// Validar WhatsApp (se informado)
	if req.WhatsApp != "" && !utils.IsValidPhone(req.WhatsApp) {
		return errors.New("whatsapp inválido")
	}

	// Validar estado (se informado)
	if req.AddressState != "" {
		if !utils.IsValidState(req.AddressState) {
			return errors.New("UF inválida")
		}
	}

	return nil
}
