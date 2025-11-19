package repository

import (
	"github.com/Leganyst/avitoTrainee/internal/model"
	"gorm.io/gorm"
)

type GormTeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *GormTeamRepository {
	return &GormTeamRepository{db}
}

func (r *GormTeamRepository) CreateTeam(team *model.Team) error {
	return r.db.Create(team).Error
}

func (r *GormTeamRepository) GetTeamByName(name string) (*model.Team, error) {
	var team model.Team
	err := r.db.Preload("Users").Where("name = ?", name).First(&team).Error
	return &team, err
}

func (r *GormTeamRepository) TeamExists(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Team{}).Where("name = ?").Count(&count).Error
	return count > 0, err
}
