package repository

import (
	"github.com/Leganyst/avitoTrainee/internal/model"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db}
}

func (r *GormUserRepository) CreateOrUpdate(user *model.User) error {
	return r.db.
		Where("user_id = ?", user.UserID).
		Assign(user).
		FirstOrCreate(user).Error
}

func (r *GormUserRepository) SetActive(userID string, active bool) (*model.User, error) {
	var user model.User
	err := r.db.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	user.IsActive = active
	return &user, r.db.Save(&user).Error
}

func (r *GormUserRepository) GetUsersByTeam(teamID uint) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("team_id = ?", teamID).Find(&users).Error
	return users, err
}

func (r *GormUserRepository) GetActiveUsersByTeam(teamID uint) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("team_id = ? AND is_active = true", teamID).Find(&users).Error
	return users, err
}
