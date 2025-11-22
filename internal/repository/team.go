package repository

import (
	"errors"

	"github.com/Leganyst/avitoTrainee/internal/config"
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
	if err := r.db.Create(team).Error; err != nil {
		config.Logger().Errorw("db create team failed", "team", team, "error", err)
		return err
	}
	config.Logger().Debugw("db team created", "team", team)
	return nil
}

func (r *GormTeamRepository) GetTeamByName(name string) (*model.Team, error) {
	var team model.Team
	if err := r.db.Preload("Users").Where("name = ?", name).First(&team).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config.Logger().Warnw("db team not found", "team_name", name)
			return nil, repoerrs.ErrNotFound
		}
		config.Logger().Errorw("db get team failed", "team_name", name, "error", err)
		return nil, err
	}
	config.Logger().Debugw("db team loaded", "team_name", name, "members", len(team.Users))
	return &team, nil
}

func (r *GormTeamRepository) TeamExists(name string) (bool, error) {
	var count int64
	err := r.db.
		Model(&model.Team{}).
		Where("name = ?", name).
		Count(&count).Error
	if err != nil {
		config.Logger().Errorw("db team exists check failed", "team_name", name, "error", err)
		return false, err
	}
	config.Logger().Debugw("db team exists check", "team_name", name, "count", count)
	return count > 0, err
}
