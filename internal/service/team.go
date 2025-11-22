package service

import (
	"errors"

	"github.com/Leganyst/avitoTrainee/internal/config"
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
	logger := config.Logger()
	exists, err := s.teamRepo.TeamExists(teamName)
	if err != nil {
		logger.Errorw("team exists check failed", "team_name", teamName, "error", err)
		return nil, err
	}
	logger.Debugw("team exists check result", "team_name", teamName, "exists", exists)

	var team *model.Team
	if exists {
		logger.Warnw("team already exists", "team_name", teamName)
		return nil, errs.ErrTeamExists
	} else {
		team = &model.Team{
			Name: teamName,
		}
		if err := s.teamRepo.CreateTeam(team); err != nil {
			if errors.Is(err, repoerrs.ErrDuplicate) {
				logger.Warnw("team create duplicate", "team_name", teamName)
				return nil, errs.ErrTeamExists
			}
			logger.Errorw("team create failed", "team_name", teamName, "error", err)
			return nil, err
		}
	}

	updatedUsers := make([]model.User, 0, len(members))
	for _, m := range members {
		user := m
		user.TeamID = team.ID

		if err := s.userRepo.CreateOrUpdate(&user); err != nil {
			logger.Errorw("create or update member failed", "team_name", teamName, "user_id", user.UserID, "error", err)
			return nil, err
		}

		logger.Debugw("team member processed", "team_name", teamName, "user_id", user.UserID)
		updatedUsers = append(updatedUsers, user)
	}

	team.Users = updatedUsers
	logger.Infow("team created", "team_name", teamName, "members", len(updatedUsers))
	return team, nil
}

func (s *teamService) GetTeam(name string) (*model.Team, error) {
	logger := config.Logger()
	team, err := s.teamRepo.GetTeamByName(name)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("team not found", "team_name", name)
			return nil, errs.ErrTeamNotFound
		}
		logger.Errorw("get team failed", "team_name", name, "error", err)
		return nil, err
	}
	logger.Infow("team fetched", "team_name", name, "members", len(team.Users))
	return team, nil
}
