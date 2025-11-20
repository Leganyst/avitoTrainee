package service

import (
	"errors"
	"math/rand"
	"time"

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
	PRService interface {
		// CreatePR создаёт PR и автоматически назначает ревьюверов согласно ТЗ.
		CreatePR(prID, name, authorID string) (*model.PullRequest, error)
		// Merge помечает PR как MERGED, операция идемпотентна.
		Merge(prID string) (*model.PullRequest, error)
		// Reassign заменяет одного ревьювера на другого из его команды.
		Reassign(prID string, oldReviewerID string) (*model.PullRequest, string, error)
	}

	prService struct {
		repo     repository.PRRepository
		userRepo repository.UserRepository
	}
)

const (
	statusOpen   = "OPEN"
	statusMerged = "MERGED"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func NewPrService(repo repository.PRRepository, userRepo repository.UserRepository) PRService {
	return &prService{repo: repo, userRepo: userRepo}
}

// CreatePR создаёт PR и разово назначает до двух случайных активных ревьюверов из команды автора.
func (s *prService) CreatePR(prID, name, authorID string) (*model.PullRequest, error) {
	author, err := s.userRepo.GetByUserID(authorID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, serviceerrs.ErrUserNotFound
		}
		return nil, err
	}

	excluded := map[uint]struct{}{author.ID: {}}
	reviewers, err := s.selectReviewers(author.TeamID, excluded, 2)
	if err != nil {
		return nil, err
	}

	pr := &model.PullRequest{
		PRID:     prID,
		Name:     name,
		Status:   statusOpen,
		AuthorID: authorID,
	}

	if err := s.repo.CreatePR(pr); err != nil {
		if errors.Is(err, repoerrs.ErrDuplicate) {
			return nil, serviceerrs.ErrPRExists
		}
		return nil, err
	}

	if len(reviewers) > 0 {
		if err := s.repo.AddReviewers(pr, reviewers); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = reviewers
	}

	return pr, nil
}

// Merge переводит PR в состояние MERGED и безопасно повторяется без побочных эффектов.
func (s *prService) Merge(prID string) (*model.PullRequest, error) {
	pr, err := s.repo.GetPRByExternalID(prID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, serviceerrs.ErrPRNotFound
		}
		return nil, err
	}

	if pr.Status == statusMerged {
		return pr, nil
	}

	pr.Status = statusMerged
	now := time.Now()
	pr.UpdatedAt = &now

	if err := s.repo.UpdatePR(pr); err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, serviceerrs.ErrPRNotFound
		}
		return nil, err
	}

	return pr, nil
}

// Reassign заменяет указанного ревьювера активным участником из его команды.
func (s *prService) Reassign(prID string, oldReviewerID string) (*model.PullRequest, string, error) {
	pr, err := s.repo.GetPRByExternalID(prID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, "", serviceerrs.ErrPRNotFound
		}
		return nil, "", err
	}

	if pr.Status == statusMerged {
		return nil, "", serviceerrs.ErrPRMerged
	}

	oldReviewer, err := s.userRepo.GetByUserID(oldReviewerID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, "", serviceerrs.ErrReviewerMissing
		}
		return nil, "", err
	}

	if !isReviewerAssigned(pr, oldReviewer.ID) {
		return nil, "", serviceerrs.ErrReviewerMissing
	}

	// Создаем map во избежание назначения ревьюером того же человека
	excluded := make(map[uint]struct{}, len(pr.AssignedReviewers)+1)
	excluded[oldReviewer.ID] = struct{}{}
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer.ID != oldReviewer.ID {
			excluded[reviewer.ID] = struct{}{}
		}
	}

	candidates, err := s.selectReviewers(oldReviewer.TeamID, excluded, 1)
	if err != nil {
		return nil, "", err
	}
	if len(candidates) == 0 {
		return nil, "", serviceerrs.ErrNoCandidates
	}
	newReviewer := candidates[0]

	if err := s.repo.ReplaceReviewer(pr, oldReviewer.ID, newReviewer.ID); err != nil {
		return nil, "", err
	}

	for i := range pr.AssignedReviewers {
		if pr.AssignedReviewers[i].ID == oldReviewer.ID {
			pr.AssignedReviewers[i] = newReviewer
			break
		}
	}

	return pr, newReviewer.UserID, nil
}

func (s *prService) selectReviewers(teamID uint, exclude map[uint]struct{}, limit int) ([]model.User, error) {
	users, err := s.userRepo.GetActiveUsersByTeam(teamID)
	if err != nil {
		return nil, err
	}

	filtered := make([]model.User, 0, len(users))
	for _, user := range users {
		if _, skip := exclude[user.ID]; skip {
			continue
		}
		filtered = append(filtered, user)
	}

	if len(filtered) == 0 {
		return nil, nil
	}

	rnd.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func isReviewerAssigned(pr *model.PullRequest, reviewerID uint) bool {
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer.ID == reviewerID {
			return true
		}
	}
	return false
}
