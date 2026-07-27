// ================================================================
// PACOTE SERVICES - ANAMNESE SERVICE
// ================================================================
// Camada de serviço responsável pela gestão de anamneses e rastreamentos.
//
// RESPONSABILIDADES:
// 1. Criar anamnese inicial
// 2. Criar rastreamentos periódicos (1 mês, 3 meses, 6 meses)
// 3. Listar anamneses por paciente
// 4. Buscar anamnese por ID
// 5. Atualizar anamnese
// 6. Acompanhamento contínuo
// ================================================================

package services

import (
	"errors"

	"cannacare-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================================================
// STRUCT ANAMNESESERVICE
// ================================================================
type AnamneseService struct {
	db *gorm.DB
}

// ================================================================
// FUNÇÃO NEWANAMNESESERVICE()
// ================================================================
func NewAnamneseService(db *gorm.DB) *AnamneseService {
	return &AnamneseService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA REQUISIÇÕES E RESPOSTAS
// ================================================================

// CreateAnamneseRequest - Dados para criar uma anamnese
type CreateAnamneseRequest struct {
	Type                string                 `json:"type" validate:"required,oneof=inicial rastreio_1_mes rastreio_3_meses rastreio_6_meses acompanhamento_continuo"`
	Symptoms            string                 `json:"symptoms" validate:"omitempty"`
	SymptomIntensity    *int                   `json:"symptom_intensity" validate:"omitempty,min=1,max=10"`
	SideEffects         string                 `json:"side_effects" validate:"omitempty"`
	SideEffectIntensity *int                   `json:"side_effect_intensity" validate:"omitempty,min=1,max=10"`
	TreatmentAdherence  string                 `json:"treatment_adherence" validate:"omitempty,oneof=alta media baixa"`
	Challenges          string                 `json:"challenges" validate:"omitempty"`
	Improvements        string                 `json:"improvements" validate:"omitempty"`
	AdditionalNotes     string                 `json:"additional_notes" validate:"omitempty"`
	Weight              *float64               `json:"weight" validate:"omitempty,min=0,max=300"`
	BloodPressure       string                 `json:"blood_pressure" validate:"omitempty"`
	HeartRate           *int                   `json:"heart_rate" validate:"omitempty,min=30,max=200"`
	ExtraResponses      map[string]interface{} `json:"extra_responses" validate:"omitempty"`
}

// UpdateAnamneseRequest - Dados para atualizar uma anamnese
type UpdateAnamneseRequest struct {
	Symptoms            string                 `json:"symptoms" validate:"omitempty"`
	SymptomIntensity    *int                   `json:"symptom_intensity" validate:"omitempty,min=1,max=10"`
	SideEffects         string                 `json:"side_effects" validate:"omitempty"`
	SideEffectIntensity *int                   `json:"side_effect_intensity" validate:"omitempty,min=1,max=10"`
	TreatmentAdherence  string                 `json:"treatment_adherence" validate:"omitempty,oneof=alta media baixa"`
	Challenges          string                 `json:"challenges" validate:"omitempty"`
	Improvements        string                 `json:"improvements" validate:"omitempty"`
	AdditionalNotes     string                 `json:"additional_notes" validate:"omitempty"`
	Weight              *float64               `json:"weight" validate:"omitempty,min=0,max=300"`
	BloodPressure       string                 `json:"blood_pressure" validate:"omitempty"`
	HeartRate           *int                   `json:"heart_rate" validate:"omitempty,min=30,max=200"`
	ExtraResponses      map[string]interface{} `json:"extra_responses" validate:"omitempty"`
}

// AnamneseResponse - Resposta com dados da anamnese
type AnamneseResponse struct {
	ID                  string                 `json:"id"`
	PatientID           string                 `json:"patient_id"`
	PatientName         string                 `json:"patient_name"`
	ResponsibleUserID   string                 `json:"responsible_user_id"`
	ResponsibleUserName string                 `json:"responsible_user_name"`
	Type                string                 `json:"type"`
	Symptoms            string                 `json:"symptoms,omitempty"`
	SymptomIntensity    *int                   `json:"symptom_intensity,omitempty"`
	SideEffects         string                 `json:"side_effects,omitempty"`
	SideEffectIntensity *int                   `json:"side_effect_intensity,omitempty"`
	TreatmentAdherence  string                 `json:"treatment_adherence,omitempty"`
	Challenges          string                 `json:"challenges,omitempty"`
	Improvements        string                 `json:"improvements,omitempty"`
	AdditionalNotes     string                 `json:"additional_notes,omitempty"`
	Weight              *float64               `json:"weight,omitempty"`
	BloodPressure       string                 `json:"blood_pressure,omitempty"`
	HeartRate           *int                   `json:"heart_rate,omitempty"`
	ExtraResponses      map[string]interface{} `json:"extra_responses,omitempty"`
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
}

// ListAnamneseRequest - Filtros para listagem
type ListAnamneseRequest struct {
	PatientID string `json:"patient_id" query:"patient_id"`
	Type      string `json:"type" query:"type"`
	Page      int    `json:"page" query:"page"`
	Limit     int    `json:"limit" query:"limit"`
}

// ================================================================
// FUNÇÃO CREATE()
// ================================================================
// Cria uma nova anamnese para um paciente
func (s *AnamneseService) Create(patientID uuid.UUID, userID uuid.UUID, req CreateAnamneseRequest) (*AnamneseResponse, error) {
	// --- 1. Validar paciente ---
	var patient models.Patient
	if err := s.db.Where("id = ?", patientID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("paciente não encontrado")
		}
		return nil, err
	}

	// --- 2. Validar tipo de anamnese ---
	validTypes := map[string]bool{
		"inicial":                 true,
		"rastreio_1_mes":          true,
		"rastreio_3_meses":        true,
		"rastreio_6_meses":        true,
		"acompanhamento_continuo": true,
	}
	if !validTypes[req.Type] {
		return nil, errors.New("tipo de anamnese inválido")
	}

	// --- 3. Verificar se já existe anamnese inicial para este paciente ---
	if req.Type == "inicial" {
		var existing models.Anamnese
		err := s.db.Where("patient_id = ? AND type = ?", patientID, "inicial").First(&existing).Error
		if err == nil {
			return nil, errors.New("paciente já possui anamnese inicial")
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	// --- 4. Validar adesão ---
	if req.TreatmentAdherence != "" {
		if req.TreatmentAdherence != "alta" && req.TreatmentAdherence != "media" && req.TreatmentAdherence != "baixa" {
			return nil, errors.New("adesão ao tratamento deve ser 'alta', 'media' ou 'baixa'")
		}
	}

	// --- 5. Validar intensidades ---
	if req.SymptomIntensity != nil && (*req.SymptomIntensity < 1 || *req.SymptomIntensity > 10) {
		return nil, errors.New("intensidade dos sintomas deve ser entre 1 e 10")
	}
	if req.SideEffectIntensity != nil && (*req.SideEffectIntensity < 1 || *req.SideEffectIntensity > 10) {
		return nil, errors.New("intensidade dos efeitos colaterais deve ser entre 1 e 10")
	}

	// --- 6. Validar dados clínicos ---
	if req.Weight != nil && (*req.Weight < 0 || *req.Weight > 300) {
		return nil, errors.New("peso deve ser entre 0 e 300 kg")
	}
	if req.HeartRate != nil && (*req.HeartRate < 30 || *req.HeartRate > 200) {
		return nil, errors.New("frequência cardíaca deve ser entre 30 e 200 bpm")
	}

	// --- 7. Criar anamnese ---
	anamnese := &models.Anamnese{
		PatientID:           patientID,
		ResponsibleUserID:   userID,
		Type:                req.Type,
		Symptoms:            req.Symptoms,
		SymptomIntensity:    req.SymptomIntensity,
		SideEffects:         req.SideEffects,
		SideEffectIntensity: req.SideEffectIntensity,
		TreatmentAdherence:  req.TreatmentAdherence,
		Challenges:          req.Challenges,
		Improvements:        req.Improvements,
		AdditionalNotes:     req.AdditionalNotes,
		Weight:              req.Weight,
		BloodPressure:       req.BloodPressure,
		HeartRate:           req.HeartRate,
		ExtraResponses:      req.ExtraResponses,
	}

	if err := s.db.Create(anamnese).Error; err != nil {
		return nil, err
	}

	// ============================================================
	// 🔧 CORREÇÃO: Recarregar a anamnese com os relacionamentos
	// ============================================================
	var createdAnamnese models.Anamnese
	if err := s.db.Preload("Patient").Preload("ResponsibleUser").
		Where("id = ?", anamnese.ID).First(&createdAnamnese).Error; err != nil {
		return nil, err
	}

	return s.toAnamneseResponse(&createdAnamnese), nil
}

// ================================================================
// FUNÇÃO GETBYID()
// ================================================================
// Busca uma anamnese pelo ID
func (s *AnamneseService) GetByID(id uuid.UUID) (*AnamneseResponse, error) {
	var anamnese models.Anamnese
	if err := s.db.Preload("Patient").Preload("ResponsibleUser").
		Where("id = ?", id).First(&anamnese).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("anamnese não encontrada")
		}
		return nil, err
	}
	return s.toAnamneseResponse(&anamnese), nil
}

// ================================================================
// FUNÇÃO GETBYPATIENT()
// ================================================================
// Lista todas as anamneses de um paciente
func (s *AnamneseService) GetByPatient(patientID uuid.UUID) ([]AnamneseResponse, error) {
	var anamneses []models.Anamnese
	if err := s.db.Preload("Patient").Preload("ResponsibleUser").
		Where("patient_id = ?", patientID).Order("created_at DESC").
		Find(&anamneses).Error; err != nil {
		return nil, err
	}

	var responses []AnamneseResponse
	for _, a := range anamneses {
		responses = append(responses, *s.toAnamneseResponse(&a))
	}

	return responses, nil
}

// ================================================================
// FUNÇÃO LIST()
// ================================================================
// Lista anamneses com filtros e paginação
func (s *AnamneseService) List(req ListAnamneseRequest) ([]AnamneseResponse, int64, error) {
	// --- 1. Definir paginação ---
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	// --- 2. Construir query ---
	query := s.db.Model(&models.Anamnese{})

	if req.PatientID != "" {
		patientID, err := uuid.Parse(req.PatientID)
		if err != nil {
			return nil, 0, errors.New("ID do paciente inválido")
		}
		query = query.Where("patient_id = ?", patientID)
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	// --- 3. Contar total ---
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// --- 4. Buscar com paginação ---
	var anamneses []models.Anamnese
	if err := query.Preload("Patient").Preload("ResponsibleUser").
		Offset(offset).Limit(req.Limit).Order("created_at DESC").
		Find(&anamneses).Error; err != nil {
		return nil, 0, err
	}

	// --- 5. Converter para resposta ---
	var responses []AnamneseResponse
	for _, a := range anamneses {
		responses = append(responses, *s.toAnamneseResponse(&a))
	}

	return responses, total, nil
}

// ================================================================
// FUNÇÃO UPDATE()
// ================================================================
// Atualiza uma anamnese existente
func (s *AnamneseService) Update(id uuid.UUID, req UpdateAnamneseRequest) (*AnamneseResponse, error) {
	var anamnese models.Anamnese
	if err := s.db.Where("id = ?", id).First(&anamnese).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("anamnese não encontrada")
		}
		return nil, err
	}

	// --- 1. Atualizar campos ---
	if req.Symptoms != "" {
		anamnese.Symptoms = req.Symptoms
	}
	if req.SymptomIntensity != nil {
		if *req.SymptomIntensity < 1 || *req.SymptomIntensity > 10 {
			return nil, errors.New("intensidade dos sintomas deve ser entre 1 e 10")
		}
		anamnese.SymptomIntensity = req.SymptomIntensity
	}
	if req.SideEffects != "" {
		anamnese.SideEffects = req.SideEffects
	}
	if req.SideEffectIntensity != nil {
		if *req.SideEffectIntensity < 1 || *req.SideEffectIntensity > 10 {
			return nil, errors.New("intensidade dos efeitos colaterais deve ser entre 1 e 10")
		}
		anamnese.SideEffectIntensity = req.SideEffectIntensity
	}
	if req.TreatmentAdherence != "" {
		if req.TreatmentAdherence != "alta" && req.TreatmentAdherence != "media" && req.TreatmentAdherence != "baixa" {
			return nil, errors.New("adesão ao tratamento deve ser 'alta', 'media' ou 'baixa'")
		}
		anamnese.TreatmentAdherence = req.TreatmentAdherence
	}
	if req.Challenges != "" {
		anamnese.Challenges = req.Challenges
	}
	if req.Improvements != "" {
		anamnese.Improvements = req.Improvements
	}
	if req.AdditionalNotes != "" {
		anamnese.AdditionalNotes = req.AdditionalNotes
	}
	if req.Weight != nil {
		if *req.Weight < 0 || *req.Weight > 300 {
			return nil, errors.New("peso deve ser entre 0 e 300 kg")
		}
		anamnese.Weight = req.Weight
	}
	if req.BloodPressure != "" {
		anamnese.BloodPressure = req.BloodPressure
	}
	if req.HeartRate != nil {
		if *req.HeartRate < 30 || *req.HeartRate > 200 {
			return nil, errors.New("frequência cardíaca deve ser entre 30 e 200 bpm")
		}
		anamnese.HeartRate = req.HeartRate
	}
	if req.ExtraResponses != nil {
		anamnese.ExtraResponses = req.ExtraResponses
	}

	if err := s.db.Save(&anamnese).Error; err != nil {
		return nil, err
	}

	return s.toAnamneseResponse(&anamnese), nil
}

// ================================================================
// FUNÇÃO DELETE()
// ================================================================
// Remove uma anamnese (soft delete)
func (s *AnamneseService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.Anamnese{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("anamnese não encontrada")
	}
	return nil
}

// ================================================================
// FUNÇÕES AUXILIARES
// ================================================================

// toAnamneseResponse - Converte models.Anamnese para AnamneseResponse
func (s *AnamneseService) toAnamneseResponse(a *models.Anamnese) *AnamneseResponse {
	// ============================================================
	// 🔧 CORREÇÃO: Verificar se os relacionamentos existem
	// ============================================================
	patientName := ""
	if a.Patient != nil && a.Patient.ID != uuid.Nil {
		patientName = a.Patient.FullName
	}

	responsibleUserName := ""
	if a.ResponsibleUser != nil && a.ResponsibleUser.ID != uuid.Nil {
		responsibleUserName = a.ResponsibleUser.Name
	}

	return &AnamneseResponse{
		ID:                  a.ID.String(),
		PatientID:           a.PatientID.String(),
		PatientName:         patientName,
		ResponsibleUserID:   a.ResponsibleUserID.String(),
		ResponsibleUserName: responsibleUserName,
		Type:                a.Type,
		Symptoms:            a.Symptoms,
		SymptomIntensity:    a.SymptomIntensity,
		SideEffects:         a.SideEffects,
		SideEffectIntensity: a.SideEffectIntensity,
		TreatmentAdherence:  a.TreatmentAdherence,
		Challenges:          a.Challenges,
		Improvements:        a.Improvements,
		AdditionalNotes:     a.AdditionalNotes,
		Weight:              a.Weight,
		BloodPressure:       a.BloodPressure,
		HeartRate:           a.HeartRate,
		ExtraResponses:      a.ExtraResponses,
		CreatedAt:           a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           a.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}