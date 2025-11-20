package service

import (
	"errors"

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
	user, err := s.userRepo.SetActive(userID, active)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, serviceerrs.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *userService) GetUserByID(userID string) (*model.User, error) {
	user, err := s.userRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, serviceerrs.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUserReviews(userID string) ([]model.PullRequest, error) {
	user, err := s.userRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, serviceerrs.ErrUserNotFound
		}
		return nil, err
	}

	return s.prRepo.GetPRsWhereReviewer(user.ID)
}
