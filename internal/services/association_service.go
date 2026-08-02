package services

import (
	"errors"
	"cannacare-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssociationService struct {
	db *gorm.DB
}

func NewAssociationService(db *gorm.DB) *AssociationService {
	return &AssociationService{db: db}
}

// ================================================================
// VERIFICAR LIMITE DE USUÁRIOS
// ================================================================
func (s *AssociationService) CheckUserLimit(associationID uuid.UUID) error {
	var association models.Association
	if err := s.db.Where("id = ?", associationID).First(&association).Error; err != nil {
		return err
	}

	var count int64
	if err := s.db.Model(&models.User{}).Where("association_id = ?", associationID).Count(&count).Error; err != nil {
		return err
	}

	// Verificar se atingiu o limite
	if int(count) >= association.UserLimit {
		return errors.New("limite de usuários do plano atingido. Faça upgrade para adicionar mais usuários")
	}

	return nil
}

// ================================================================
// OBTER LIMITE DE USUÁRIOS
// ================================================================
func (s *AssociationService) GetUserLimit(associationID uuid.UUID) (int, error) {
	var association models.Association
	if err := s.db.Where("id = ?", associationID).First(&association).Error; err != nil {
		return 0, err
	}
	return association.UserLimit, nil
}