package repository

import (
	"errors"

	"github.com/Leganyst/avitoTrainee/internal/config"
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
		BulkDeactivate(teamID uint, userIDs []string) ([]model.User, error)
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
			config.Logger().Warnw("db user duplicate", "user_id", user.UserID)
			return repoerrs.ErrDuplicate
		}
		config.Logger().Errorw("db create/update user failed", "user_id", user.UserID, "error", err)
		return err
	}
	config.Logger().Debugw("db user upserted", "user_id", user.UserID, "team_id", user.TeamID)
	return nil
}

func (r *GormUserRepository) SetActive(userID string, active bool) (*model.User, error) {
	var user model.User
	if err := r.db.Where("user_id = ?", userID).
		Preload("Team").
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config.Logger().Warnw("db user not found for set active", "user_id", userID)
			return nil, repoerrs.ErrNotFound
		}
		config.Logger().Errorw("db get user for set active failed", "user_id", userID, "error", err)
		return nil, err
	}

	user.IsActive = active
	if err := r.db.Save(&user).Error; err != nil {
		config.Logger().Errorw("db save user active failed", "user_id", userID, "error", err)
		return nil, err
	}
	config.Logger().Debugw("db user active updated", "user_id", userID, "is_active", active)
	return &user, nil
}

func (r *GormUserRepository) GetUsersByTeam(teamID uint) ([]model.User, error) {
	var users []model.User
	err := r.db.
		Where("team_id = ?", teamID).
		Find(&users).Error
	if err != nil {
		config.Logger().Errorw("db get users by team failed", "team_id", teamID, "error", err)
		return nil, err
	}
	config.Logger().Debugw("db users by team loaded", "team_id", teamID, "count", len(users))
	return users, err
}

func (r *GormUserRepository) GetActiveUsersByTeam(teamID uint) ([]model.User, error) {
	var users []model.User
	err := r.db.
		Where("team_id = ? AND is_active = true", teamID).
		Find(&users).Error
	if err != nil {
		config.Logger().Errorw("db get active users failed", "team_id", teamID, "error", err)
		return nil, err
	}
	config.Logger().Debugw("db active users loaded", "team_id", teamID, "count", len(users))
	return users, err
}

func (r *GormUserRepository) GetByUserID(userID string) (*model.User, error) {
	var user model.User
	if err := r.db.
		Where("user_id = ?", userID).
		Preload("Team").
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config.Logger().Warnw("db user not found", "user_id", userID)
			return nil, repoerrs.ErrNotFound
		}
		config.Logger().Errorw("db get user failed", "user_id", userID, "error", err)
		return nil, err
	}
	config.Logger().Debugw("db user loaded", "user_id", userID, "team_id", user.TeamID)
	return &user, nil
}

func (r *GormUserRepository) BulkDeactivate(teamID uint, userIDs []string) ([]model.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	var users []model.User
	if err := r.db.
		Where("team_id = ? AND user_id IN ? AND is_active = true", teamID, userIDs).
		Find(&users).Error; err != nil {
		config.Logger().Errorw("db find users for bulk deactivate failed", "team_id", teamID, "user_ids", userIDs, "error", err)
		return nil, err
	}
	if len(users) == 0 {
		return nil, repoerrs.ErrNotFound
	}

	ids := make([]uint, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}

	if err := r.db.Model(&model.User{}).
		Where("id IN ?", ids).
		Update("is_active", false).Error; err != nil {
		config.Logger().Errorw("db bulk deactivate failed", "ids", ids, "error", err)
		return nil, err
	}

	config.Logger().Infow("db bulk deactivate completed", "count", len(users), "team_id", teamID)
	return users, nil
}
