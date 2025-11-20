package repository

import (
	"errors"

	"github.com/Leganyst/avitoTrainee/internal/model"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
	"gorm.io/gorm"
)

type (
	UserRepository interface {
		CreateOrUpdate(user *model.User) error
		GetByUserID(userID string) (*model.User, error)
		GetUsersByTeam(teamID uint) ([]model.User, error)
		SetActive(userID string, active bool) (*model.User, error)

		GetActiveUsersByTeam(teamID uint) ([]model.User, error)
	}

	GormUserRepository struct {
		db *gorm.DB
	}
)

func NewUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db}
}

func (r *GormUserRepository) CreateOrUpdate(user *model.User) error {
	if err := r.db.
		Where("user_id = ?", user.UserID).
		Assign(user).
		FirstOrCreate(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return repoerrs.ErrDuplicate
		}
		return err
	}
	return nil
}

func (r *GormUserRepository) SetActive(userID string, active bool) (*model.User, error) {
	var user model.User
	if err := r.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repoerrs.ErrNotFound
		}
		return nil, err
	}

	user.IsActive = active
	if err := r.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) GetUsersByTeam(teamID uint) ([]model.User, error) {
	var users []model.User
	err := r.db.
		Where("team_id = ?", teamID).
		Find(&users).Error
	return users, err
}

func (r *GormUserRepository) GetActiveUsersByTeam(teamID uint) ([]model.User, error) {
	var users []model.User
	err := r.db.
		Where("team_id = ? AND is_active = true", teamID).
		Find(&users).Error
	return users, err
}

func (r *GormUserRepository) GetByUserID(userID string) (*model.User, error) {
	var user model.User
	if err := r.db.
		Where("user_id = ?", userID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repoerrs.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
