package service

import (
	"errors"

	"github.com/Leganyst/avitoTrainee/internal/model"
	"github.com/Leganyst/avitoTrainee/internal/repository"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
	"github.com/Leganyst/avitoTrainee/internal/service/errs"
)

/*
Использование ORM моделей в качестве доменных - это не совсем чистая архитектура проекта.
Однако, ТЗ маленькое и детальных требований по организации разделения слоев - нет.
Следовательно, решение об использовании ORM моделей в качестве доменных принято только мной.
Это сократит общее количество кодовой базы и ускорит разработку конечного решения согласно ТЗ.
При гораздо длительной разработке и поддержке сервиса нужно определять отдельные доменные модели,
использующиеся в ServiceLayer

Итого, в рамках выполняемой работы достаточно использовать model.* модели (ORM-модели)

*/

type (
	TeamService interface {
		CreateTeam(teamName string, members []model.User) (*model.Team, error)
		GetTeam(name string) (*model.Team, error)
	}

	teamService struct {
		teamRepo repository.TeamRepository
		userRepo repository.UserRepository
	}
)

func NewTeamService(
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *teamService) CreateTeam(teamName string, members []model.User) (*model.Team, error) {
	exists, err := s.teamRepo.TeamExists(teamName)
	if err != nil {
		return nil, err
	}

	var team *model.Team
	if exists {
		return nil, errs.ErrTeamExists
	} else {
		team = &model.Team{
			Name: teamName,
		}
		if err := s.teamRepo.CreateTeam(team); err != nil {
			if errors.Is(err, repoerrs.ErrDuplicate) {
				return nil, errs.ErrTeamExists
			}
			return nil, err
		}
	}

	updatedUsers := make([]model.User, 0, len(members))
	for _, m := range members {
		user := m
		user.TeamID = team.ID

		if err := s.userRepo.CreateOrUpdate(&user); err != nil {
			return nil, err
		}

		updatedUsers = append(updatedUsers, user)
	}

	team.Users = updatedUsers
	return team, nil
}

func (s *teamService) GetTeam(name string) (*model.Team, error) {
	team, err := s.teamRepo.GetTeamByName(name)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, errs.ErrTeamNotFound
		}
		return nil, err
	}
	return team, nil
}
