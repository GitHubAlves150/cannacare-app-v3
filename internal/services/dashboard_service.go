// ================================================================
// PACOTE SERVICES - DASHBOARD SERVICE
// ================================================================
// ⚠️ CORRIGIDO: nenhuma função aqui filtrava por associationID —
// toda "Visão Geral" mostrava dados somados de TODAS as associações
// do banco. Agora todas recebem e filtram por associationID.
// ================================================================

package services

import (
	"time"

	"cannacare-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

type DashboardOverview struct {
	Patients  PatientStats   `json:"patients"`
	Doctors   DoctorStats    `json:"doctors"`
	Orders    OrderStats     `json:"orders"`
	Financial FinancialStats `json:"financial"`
	Stock     StockStats     `json:"stock"`
	UpdatedAt string         `json:"updated_at"`
}

type PatientStats struct {
	Total        int64 `json:"total"`
	Approved     int64 `json:"approved"`
	Pending      int64 `json:"pending"`
	InAnalysis   int64 `json:"in_analysis"`
	Negated      int64 `json:"negated"`
	Social       int64 `json:"social"`
	NewThisMonth int64 `json:"new_this_month"`
}

type DoctorStats struct {
	Total            int64  `json:"total"`
	Active           int64  `json:"active"`
	Inactive         int64  `json:"inactive"`
	TopDoctor        string `json:"top_doctor"`
	TopPrescriptions int64  `json:"top_prescriptions"`
}

type OrderStats struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Separated int64 `json:"separated"`
	Dispensa  int64 `json:"dispensa"`
	Correio   int64 `json:"correio"`
	Delivered int64 `json:"delivered"`
	Canceled  int64 `json:"canceled"`
	ThisMonth int64 `json:"this_month"`
}

type FinancialStats struct {
	TotalRevenue         float64 `json:"total_revenue"`
	MonthRevenue         float64 `json:"month_revenue"`
	PendingPayments      float64 `json:"pending_payments"`
	OverdueSubscriptions int64   `json:"overdue_subscriptions"`
	ActiveSubscriptions  int64   `json:"active_subscriptions"`
}

type StockStats struct {
	TotalProducts int64 `json:"total_products"`
	TotalLots     int64 `json:"total_lots"`
	TotalQuantity int64 `json:"total_quantity"`
	LowStockItems int64 `json:"low_stock_items"`
	ExpiringItems int64 `json:"expiring_items"`
}

// ================================================================
// FUNÇÃO GETOVERVIEW()
// ================================================================
func (s *DashboardService) GetOverview(associationID uuid.UUID) (*DashboardOverview, error) {
	overview := &DashboardOverview{
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.getPatientStats(associationID, overview); err != nil {
		return nil, err
	}
	if err := s.getDoctorStats(associationID, overview); err != nil {
		return nil, err
	}
	if err := s.getOrderStats(associationID, overview); err != nil {
		return nil, err
	}
	if err := s.getFinancialStats(associationID, overview); err != nil {
		return nil, err
	}
	if err := s.getStockStats(associationID, overview); err != nil {
		return nil, err
	}

	return overview, nil
}

// ================================================================
// FUNÇÕES AUXILIARES PARA ESTATÍSTICAS (todas filtram por associação)
// ================================================================

func (s *DashboardService) getPatientStats(associationID uuid.UUID, overview *DashboardOverview) error {
	stats := &overview.Patients
	base := s.db.Model(&models.Patient{}).Where("association_id = ?", associationID)

	base.Session(&gorm.Session{}).Count(&stats.Total)
	base.Session(&gorm.Session{}).Where("status = ?", "aprovado").Count(&stats.Approved)
	base.Session(&gorm.Session{}).Where("status = ?", "pendente_documentacao").Count(&stats.Pending)
	base.Session(&gorm.Session{}).Where("status = ?", "em_analise").Count(&stats.InAnalysis)
	base.Session(&gorm.Session{}).Where("status = ?", "negado").Count(&stats.Negated)
	base.Session(&gorm.Session{}).Where("is_social_patient = ?", true).Count(&stats.Social)

	startOfMonth := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	base.Session(&gorm.Session{}).Where("created_at >= ?", startOfMonth).Count(&stats.NewThisMonth)

	return nil
}

func (s *DashboardService) getDoctorStats(associationID uuid.UUID, overview *DashboardOverview) error {
	stats := &overview.Doctors
	base := s.db.Model(&models.Doctor{}).Where("association_id = ?", associationID)

	base.Session(&gorm.Session{}).Count(&stats.Total)
	base.Session(&gorm.Session{}).Where("is_active = ?", true).Count(&stats.Active)
	base.Session(&gorm.Session{}).Where("is_active = ?", false).Count(&stats.Inactive)

	// ⚠️ Também corrigido: os nomes de coluna não batiam com os campos
	// do struct (a query nunca preenchia TopPrescriptions de verdade).
	var result struct {
		DoctorName         string
		TotalPrescriptions int64
	}
	s.db.Table("vw_top_doctors").
		Select("doctor_name, total_prescriptions").
		Where("association_id = ?", associationID).
		Order("total_prescriptions DESC").
		Limit(1).
		Scan(&result)

	stats.TopDoctor = result.DoctorName
	stats.TopPrescriptions = result.TotalPrescriptions

	return nil
}

func (s *DashboardService) getOrderStats(associationID uuid.UUID, overview *DashboardOverview) error {
	stats := &overview.Orders
	base := s.db.Model(&models.Order{}).Where("association_id = ?", associationID)

	base.Session(&gorm.Session{}).Count(&stats.Total)
	base.Session(&gorm.Session{}).Where("status = ?", "pendente").Count(&stats.Pending)
	base.Session(&gorm.Session{}).Where("status = ?", "separado").Count(&stats.Separated)
	base.Session(&gorm.Session{}).Where("status = ?", "dispensa").Count(&stats.Dispensa)
	base.Session(&gorm.Session{}).Where("status = ?", "correio").Count(&stats.Correio)
	base.Session(&gorm.Session{}).Where("status = ?", "entregue").Count(&stats.Delivered)
	base.Session(&gorm.Session{}).Where("status = ?", "cancelado").Count(&stats.Canceled)

	startOfMonth := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	base.Session(&gorm.Session{}).Where("created_at >= ?", startOfMonth).Count(&stats.ThisMonth)

	return nil
}

func (s *DashboardService) getFinancialStats(associationID uuid.UUID, overview *DashboardOverview) error {
	stats := &overview.Financial

	s.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.association_id = ? AND payments.status = ?", associationID, "pago").
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&stats.TotalRevenue)

	startOfMonth := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	s.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.association_id = ? AND payments.status = ? AND payments.paid_at >= ?", associationID, "pago", startOfMonth).
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&stats.MonthRevenue)

	s.db.Model(&models.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.association_id = ? AND payments.status IN (?)", associationID, []string{"pendente", "recusado"}).
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&stats.PendingPayments)

	s.db.Model(&models.Subscription{}).
		Joins("JOIN patients ON patients.id = subscriptions.patient_id").
		Where("patients.association_id = ? AND subscriptions.status = ?", associationID, "atrasado").
		Count(&stats.OverdueSubscriptions)

	s.db.Model(&models.Subscription{}).
		Joins("JOIN patients ON patients.id = subscriptions.patient_id").
		Where("patients.association_id = ? AND subscriptions.status IN (?)", associationID, []string{"pago", "pendente"}).
		Count(&stats.ActiveSubscriptions)

	return nil
}

func (s *DashboardService) getStockStats(associationID uuid.UUID, overview *DashboardOverview) error {
	stats := &overview.Stock

	s.db.Model(&models.Product{}).Where("association_id = ?", associationID).Count(&stats.TotalProducts)
	s.db.Model(&models.ProductLot{}).Where("association_id = ?", associationID).Count(&stats.TotalLots)

	s.db.Model(&models.ProductLot{}).
		Where("association_id = ?", associationID).
		Select("COALESCE(SUM(current_quantity), 0)").
		Scan(&stats.TotalQuantity)

	var lowStockCount int64
	s.db.Table("vw_low_stock").Where("association_id = ?", associationID).Count(&lowStockCount)
	stats.LowStockItems = lowStockCount

	var expiringCount int64
	s.db.Model(&models.ProductLot{}).
		Where("association_id = ?", associationID).
		Where("expiration_date BETWEEN ? AND ?", time.Now(), time.Now().AddDate(0, 0, 30)).
		Where("current_quantity > 0").
		Count(&expiringCount)
	stats.ExpiringItems = expiringCount

	return nil
}

// ================================================================
// FUNÇÕES PARA RELATÓRIOS ESPECÍFICOS (todas filtram por associação)
// ================================================================

func (s *DashboardService) GetPatientReport(associationID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_patient_dashboard").Where("association_id = ?", associationID).Find(&results).Error
	return results, err
}

func (s *DashboardService) GetExpiredPrescriptionsReport(associationID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_expired_prescriptions").Where("association_id = ?", associationID).Find(&results).Error
	return results, err
}

func (s *DashboardService) GetTopDoctorsReport(associationID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_top_doctors").Where("association_id = ?", associationID).Find(&results).Error
	return results, err
}

func (s *DashboardService) GetLowStockReport(associationID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("vw_low_stock").Where("association_id = ?", associationID).Find(&results).Error
	return results, err
}