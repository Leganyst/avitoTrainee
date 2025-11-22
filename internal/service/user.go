package service

import (
	"errors"

	"github.com/Leganyst/avitoTrainee/internal/config"
	"github.com/Leganyst/avitoTrainee/internal/model"
	"github.com/Leganyst/avitoTrainee/internal/repository"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
	serviceerrs "github.com/Leganyst/avitoTrainee/internal/service/errs"
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
	UserService interface {
		SetActive(userID string, active bool) (*model.User, error)
		GetUserByID(userID string) (*model.User, error)
		GetUserReviews(userID string) ([]model.PullRequest, error)
	}

	userService struct {
		userRepo repository.UserRepository
		prRepo   repository.PRRepository
	}
)

func NewUserService(userRepo repository.UserRepository, prRepo repository.PRRepository) UserService {
	return &userService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *userService) SetActive(userID string, active bool) (*model.User, error) {
	logger := config.Logger()
	user, err := s.userRepo.SetActive(userID, active)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("set active user not found", "user_id", userID)
			return nil, serviceerrs.ErrUserNotFound
		}
		logger.Errorw("set active failed", "user_id", userID, "error", err)
		return nil, err
	}

	logger.Debugw("user entity after set active", "user", user)
	logger.Infow("user activity updated", "user_id", userID, "is_active", active)
	return user, nil
}

func (s *userService) GetUserByID(userID string) (*model.User, error) {
	logger := config.Logger()
	user, err := s.userRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("user not found", "user_id", userID)
			return nil, serviceerrs.ErrUserNotFound
		}
		logger.Errorw("get user failed", "user_id", userID, "error", err)
		return nil, err
	}
	logger.Debugw("user entity fetched", "user", user)
	logger.Infow("user fetched", "user_id", userID, "team_id", user.TeamID)
	return user, nil
}

func (s *userService) GetUserReviews(userID string) ([]model.PullRequest, error) {
	logger := config.Logger()
	user, err := s.userRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("reviews user not found", "user_id", userID)
			return nil, serviceerrs.ErrUserNotFound
		}
		logger.Errorw("get user before reviews failed", "user_id", userID, "error", err)
		return nil, err
	}

	prs, err := s.prRepo.GetPRsWhereReviewer(user.ID)
	if err != nil {
		logger.Errorw("get reviews failed", "user_id", userID, "error", err)
		return nil, err
	}
	logger.Debugw("user reviews list", "user_id", userID, "prs", prs)
	logger.Infow("user reviews fetched", "user_id", userID, "count", len(prs))
	return prs, nil
}
