package service

import "github.com/Leganyst/avitoTrainee/internal/model"

type TeamRepository interface {
	CreateTeam(team *model.Team) error
	GetTeamByName(name string) (*model.Team, error)
	TeamExists(name string) (bool, error)
}
