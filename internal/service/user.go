package service

import "github.com/Leganyst/avitoTrainee/internal/model"

type UserRepository interface {
	CreateOrUpdate(user *model.User) error
	GetByUserID(userID string) (*model.User, error)
	GetUsersByTeam(teamID uint) ([]model.User, error)
	SetActive(userID string, active bool) (*model.User, error)

	GetActiveUsersByTeam(teamID uint) ([]model.User, error)
}
