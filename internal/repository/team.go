package repository

import (
	"errors"

	"github.com/Leganyst/avitoTrainee/internal/model"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
	"gorm.io/gorm"
)

type (
	TeamRepository interface {
		CreateTeam(team *model.Team) error
		GetTeamByName(name string) (*model.Team, error)
		TeamExists(name string) (bool, error)
	}

	GormTeamRepository struct {
		db *gorm.DB
	}
)

func NewTeamRepository(db *gorm.DB) *GormTeamRepository {
	return &GormTeamRepository{db}
}

func (r *GormTeamRepository) CreateTeam(team *model.Team) error {
	return r.db.Create(team).Error
}

func (r *GormTeamRepository) GetTeamByName(name string) (*model.Team, error) {
	var team model.Team
	if err := r.db.Preload("Users").Where("name = ?", name).First(&team).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repoerrs.ErrNotFound
		}
		return nil, err
	}
	return &team, nil
}

func (r *GormTeamRepository) TeamExists(name string) (bool, error) {
	var count int64
	err := r.db.
		Model(&model.Team{}).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}
