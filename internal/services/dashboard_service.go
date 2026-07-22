// ================================================================
// PACOTE SERVICES - DASHBOARD SERVICE
// ================================================================
// Camada de serviço responsável por gerar dados para dashboards e relatórios.
//
// RESPONSABILIDADES:
// 1. Visão geral do sistema (overview)
// 2. Estatísticas de pacientes
// 3. Estatísticas de médicos
// 4. Estatísticas de pedidos
// 5. Estatísticas financeiras
// 6. Estatísticas de estoque
// 7. Relatórios consolidados
// ================================================================

package services

import (
	"time"

	"cannacare-backend/internal/models"

	"gorm.io/gorm"
)

// ================================================================
// STRUCT DASHBOARDSERVICE
// ================================================================
type DashboardService struct {
	db *gorm.DB
}

// ================================================================
// FUNÇÃO NEWDASHBOARDSERVICE()
// ================================================================
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{
		db: db,
	}
}

// ================================================================
// STRUCTS PARA RESPOSTAS
// ================================================================

// DashboardOverview - Visão geral do sistema
type DashboardOverview struct {
	Patients    PatientStats    `json:"patients"`
	Doctors     DoctorStats     `json:"doctors"`
	Orders      OrderStats      `json:"orders"`
	Financial   FinancialStats  `json:"financial"`
	Stock       StockStats      `json:"stock"`
	UpdatedAt   string          `json:"updated_at"`
}

// PatientStats - Estatísticas de pacientes
type PatientStats struct {
	Total           int64 `json:"total"`
	Approved        int64 `json:"approved"`
	Pending         int64 `json:"pending"`
	InAnalysis      int64 `json:"in_analysis"`
	Negated         int64 `json:"negated"`
	Social          int64 `json:"social"`
	NewThisMonth    int64 `json:"new_this_month"`
}

// DoctorStats - Estatísticas de médicos
type DoctorStats struct {
	Total       int64 `json:"total"`
	Active      int64 `json:"active"`
	Inactive    int64 `json:"inactive"`
	TopDoctor   string `json:"top_doctor"`
	TopPrescriptions int64 `json:"top_prescriptions"`
}

// OrderStats - Estatísticas de pedidos
type OrderStats struct {
	Total       int64 `json:"total"`
	Pending     int64 `json:"pending"`
	Separated   int64 `json:"separated"`
	Dispensa    int64 `json:"dispensa"`
	Correio     int64 `json:"correio"`
	Delivered   int64 `json:"delivered"`
	Canceled    int64 `json:"canceled"`
	ThisMonth   int64 `json:"this_month"`
}

// FinancialStats - Estatísticas financeiras
type FinancialStats struct {
	TotalRevenue      float64 `json:"total_revenue"`
	MonthRevenue      float64 `json:"month_revenue"`
	PendingPayments   float64 `json:"pending_payments"`
	OverdueSubscriptions int64 `json:"overdue_subscriptions"`
	ActiveSubscriptions int64 `json:"active_subscriptions"`
}

// StockStats - Estatísticas de estoque
type StockStats struct {
	TotalProducts    int64 `json:"total_products"`
	TotalLots        int64 `json:"total_lots"`
	TotalQuantity    int64 `json:"total_quantity"`
	LowStockItems    int64 `json:"low_stock_items"`
	ExpiringItems    int64 `json:"expiring_items"`
}

// ================================================================
// FUNÇÃO GETOVERVIEW()
// ================================================================
// Retorna a visão geral do sistema
func (s *DashboardService) GetOverview() (*DashboardOverview, error) {
	overview := &DashboardOverview{
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// --- 1. Estatísticas de Pacientes ---
	if err := s.getPatientStats(overview); err != nil {
		return nil, err
	}

	// --- 2. Estatísticas de Médicos ---
	if err := s.getDoctorStats(overview); err != nil {
		return nil, err
	}

	// --- 3. Estatísticas de Pedidos ---
	if err := s.getOrderStats(overview); err != nil {
		return nil, err
	}

	// --- 4. Estatísticas Financeiras ---
	if err := s.getFinancialStats(overview); err != nil {
		return nil, err
	}

	// --- 5. Estatísticas de Estoque ---
	if err := s.getStockStats(overview); err != nil {
		return nil, err
	}

	return overview, nil
}

// ================================================================
// FUNÇÕES AUXILIARES PARA ESTATÍSTICAS
// ================================================================

// getPatientStats - Busca estatísticas de pacientes
func (s *DashboardService) getPatientStats(overview *DashboardOverview) error {
	stats := &overview.Patients

	// Total de pacientes
	s.db.Model(&models.Patient{}).Count(&stats.Total)

	// Pacientes por status
	s.db.Model(&models.Patient{}).Where("status = ?", "aprovado").Count(&stats.Approved)
	s.db.Model(&models.Patient{}).Where("status = ?", "pendente_documentacao").Count(&stats.Pending)
	s.db.Model(&models.Patient{}).Where("status = ?", "em_analise").Count(&stats.InAnalysis)
	s.db.Model(&models.Patient{}).Where("status = ?", "negado").Count(&stats.Negated)
	s.db.Model(&models.Patient{}).Where("is_social_patient = ?", true).Count(&stats.Social)

	// Novos pacientes este mês
	startOfMonth := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	s.db.Model(&models.Patient{}).Where("created_at >= ?", startOfMonth).Count(&stats.NewThisMonth)

	return nil
}

// getDoctorStats - Busca estatísticas de médicos
func (s *DashboardService) getDoctorStats(overview *DashboardOverview) error {
	stats := &overview.Doctors

	// Total de médicos
	s.db.Model(&models.Doctor{}).Count(&stats.Total)

	// Médicos ativos/inativos
	s.db.Model(&models.Doctor{}).Where("is_active = ?", true).Count(&stats.Active)
	s.db.Model(&models.Doctor{}).Where("is_active = ?", false).Count(&stats.Inactive)

	// Médico que mais prescreve
	var result struct {
		DoctorName string
		Count      int64
	}
	s.db.Table("vw_top_doctors").
		Select("doctor_name, total_prescriptions").
		Order("total_prescriptions DESC").
		Limit(1).
		Scan(&result)

	stats.TopDoctor = result.DoctorName
	stats.TopPrescriptions = result.Count

	return nil
}

// getOrderStats - Busca estatísticas de pedidos
func (s *DashboardService) getOrderStats(overview *DashboardOverview) error {
	stats := &overview.Orders

	// Total de pedidos
	s.db.Model(&models.Order{}).Count(&stats.Total)

	// Pedidos por status
	s.db.Model(&models.Order{}).Where("status = ?", "pendente").Count(&stats.Pending)
	s.db.Model(&models.Order{}).Where("status = ?", "separado").Count(&stats.Separated)
	s.db.Model(&models.Order{}).Where("status = ?", "dispensa").Count(&stats.Dispensa)
	s.db.Model(&models.Order{}).Where("status = ?", "correio").Count(&stats.Correio)
	s.db.Model(&models.Order{}).Where("status = ?", "entregue").Count(&stats.Delivered)
	s.db.Model(&models.Order{}).Where("status = ?", "cancelado").Count(&stats.Canceled)

	// Pedidos este mês
	startOfMonth := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	s.db.Model(&models.Order{}).Where("created_at >= ?", startOfMonth).Count(&stats.ThisMonth)

	return nil
}

// getFinancialStats - Busca estatísticas financeiras
func (s *DashboardService) getFinancialStats(overview *DashboardOverview) error {
	stats := &overview.Financial

	// Receita total (pagamentos confirmados)
	s.db.Model(&models.Payment{}).
		Where("status = ?", "pago").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalRevenue)

	// Receita deste mês
	startOfMonth := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	s.db.Model(&models.Payment{}).
		Where("status = ? AND paid_at >= ?", "pago", startOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.MonthRevenue)

	// Pagamentos pendentes
	s.db.Model(&models.Payment{}).
		Where("status IN (?)", []string{"pendente", "recusado"}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.PendingPayments)

	// Anuidades em atraso
	s.db.Model(&models.Subscription{}).
		Where("status = ?", "atrasado").
		Count(&stats.OverdueSubscriptions)

	// Anuidades ativas (pagas ou pendentes)
	s.db.Model(&models.Subscription{}).
		Where("status IN (?)", []string{"pago", "pendente"}).
		Count(&stats.ActiveSubscriptions)

	return nil
}

// getStockStats - Busca estatísticas de estoque
func (s *DashboardService) getStockStats(overview *DashboardOverview) error {
	stats := &overview.Stock

	// Total de produtos
	s.db.Model(&models.Product{}).Count(&stats.TotalProducts)

	// Total de lotes
	s.db.Model(&models.ProductLot{}).Count(&stats.TotalLots)

	// Quantidade total em estoque
	s.db.Model(&models.ProductLot{}).
		Select("COALESCE(SUM(current_quantity), 0)").
		Scan(&stats.TotalQuantity)

	// Produtos com estoque baixo
	var lowStockCount int64
	s.db.Table("vw_low_stock").Count(&lowStockCount)
	stats.LowStockItems = lowStockCount

	// Produtos com validade próxima (30 dias)
	var expiringCount int64
	s.db.Model(&models.ProductLot{}).
		Where("expiration_date BETWEEN ? AND ?", time.Now(), time.Now().AddDate(0, 0, 30)).
		Where("current_quantity > 0").
		Count(&expiringCount)
	stats.ExpiringItems = expiringCount

	return nil
}

// ================================================================
// FUNÇÕES PARA RELATÓRIOS ESPECÍFICOS
// ================================================================

// GetPatientReport - Relatório detalhado de pacientes
func (s *DashboardService) GetPatientReport() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_patient_dashboard").Find(&results).Error
	return results, err
}

// GetExpiredPrescriptionsReport - Relatório de receitas vencidas
func (s *DashboardService) GetExpiredPrescriptionsReport() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_expired_prescriptions").Find(&results).Error
	return results, err
}

// GetTopDoctorsReport - Relatório de médicos que mais prescrevem
func (s *DashboardService) GetTopDoctorsReport() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_top_doctors").Find(&results).Error
	return results, err
}

// GetLowStockReport - Relatório de produtos com estoque baixo
func (s *DashboardService) GetLowStockReport() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_low_stock").Find(&results).Error
	return results, err
}