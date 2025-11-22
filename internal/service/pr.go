package service

import (
	"errors"
	"math/rand"
	"time"

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
	logger := config.Logger()
	author, err := s.userRepo.GetByUserID(authorID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("author not found", "author_id", authorID, "pr_id", prID)
			return nil, serviceerrs.ErrUserNotFound
		}
		logger.Errorw("failed to fetch author", "author_id", authorID, "error", err)
		return nil, err
	}

	excluded := map[uint]struct{}{author.ID: {}}
	reviewers, err := s.selectReviewers(author.TeamID, excluded, 2)
	if err != nil {
		return nil, err
	}
	logger.Debugw("selected reviewers candidates", "team_id", author.TeamID, "selected", reviewers)

	pr := &model.PullRequest{
		PRID:     prID,
		Name:     name,
		Status:   statusOpen,
		AuthorID: author.ID,
		Author:   *author,
	}

	if err := s.repo.CreatePR(pr); err != nil {
		if errors.Is(err, repoerrs.ErrDuplicate) {
			logger.Warnw("PR already exists", "pr_id", prID)
			return nil, serviceerrs.ErrPRExists
		}
		logger.Errorw("failed to create PR", "pr_id", prID, "error", err)
		return nil, err
	}

	if len(reviewers) > 0 {
		if err := s.repo.AddReviewers(pr, reviewers); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = reviewers
	}

	logger.Infow("PR created", "pr_id", prID, "author", authorID, "reviewers", len(pr.AssignedReviewers))
	return pr, nil
}

// Merge переводит PR в состояние MERGED и безопасно повторяется без побочных эффектов.
func (s *prService) Merge(prID string) (*model.PullRequest, error) {
	logger := config.Logger()
	pr, err := s.repo.GetPRByExternalID(prID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("PR not found for merge", "pr_id", prID)
			return nil, serviceerrs.ErrPRNotFound
		}
		logger.Errorw("failed to fetch PR for merge", "pr_id", prID, "error", err)
		return nil, err
	}

	if pr.Status == statusMerged {
		logger.Debugw("merge called on already merged PR", "pr_id", prID)
		return pr, nil
	}

	pr.Status = statusMerged
	now := time.Now()
	pr.UpdatedAt = &now

	if err := s.repo.UpdatePR(pr); err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("PR not found on update", "pr_id", prID)
			return nil, serviceerrs.ErrPRNotFound
		}
		logger.Errorw("failed to update PR status", "pr_id", prID, "error", err)
		return nil, err
	}

	logger.Infow("PR merged", "pr_id", prID)
	return pr, nil
}

// Reassign заменяет указанного ревьювера активным участником из его команды.
func (s *prService) Reassign(prID string, oldReviewerID string) (*model.PullRequest, string, error) {
	logger := config.Logger()
	pr, err := s.repo.GetPRByExternalID(prID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("PR not found for reassign", "pr_id", prID)
			return nil, "", serviceerrs.ErrPRNotFound
		}
		logger.Errorw("failed to fetch PR for reassign", "pr_id", prID, "error", err)
		return nil, "", err
	}

	if pr.Status == statusMerged {
		logger.Warnw("reassign attempted on merged PR", "pr_id", prID)
		return nil, "", serviceerrs.ErrPRMerged
	}

	oldReviewer, err := s.userRepo.GetByUserID(oldReviewerID)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("old reviewer not found", "pr_id", prID, "user_id", oldReviewerID)
			return nil, "", serviceerrs.ErrReviewerMissing
		}
		logger.Errorw("failed to fetch old reviewer", "pr_id", prID, "user_id", oldReviewerID, "error", err)
		return nil, "", err
	}

	if !isReviewerAssigned(pr, oldReviewer.ID) {
		logger.Warnw("reviewer not assigned to PR", "pr_id", prID, "user_id", oldReviewerID)
		return nil, "", serviceerrs.ErrReviewerMissing
	}

	// Создаем map во избежание назначения ревьюером того же человека
	excluded := make(map[uint]struct{}, len(pr.AssignedReviewers)+2)
	excluded[oldReviewer.ID] = struct{}{}
	excluded[pr.AuthorID] = struct{}{}
	logger.Debugw("excluded reviewers for replacement", "pr_id", prID, "excluded_ids", excluded)

	for _, r := range pr.AssignedReviewers {
		excluded[r.ID] = struct{}{}
	}

	candidates, err := s.selectReviewers(oldReviewer.TeamID, excluded, 1)
	if err != nil {
		logger.Errorw("select replacement reviewers failed", "pr_id", prID, "error", err)
		return nil, "", err
	}
	if len(candidates) == 0 {
		logger.Warnw("no candidates for reassign", "pr_id", prID)
		return nil, "", serviceerrs.ErrNoCandidates
	}
	newReviewer := candidates[0]

	if err := s.repo.ReplaceReviewer(pr, oldReviewer.ID, newReviewer); err != nil {
		logger.Errorw("replace reviewer failed", "pr_id", prID, "old_user", oldReviewerID, "new_user", newReviewer.UserID, "error", err)
		return nil, "", err
	}

	for i := range pr.AssignedReviewers {
		if pr.AssignedReviewers[i].ID == oldReviewer.ID {
			pr.AssignedReviewers[i] = newReviewer
			break
		}
	}

	logger.Infow("reviewer replaced", "pr_id", prID, "old_user", oldReviewerID, "new_user", newReviewer.UserID)
	return pr, newReviewer.UserID, nil
}

func (s *prService) selectReviewers(teamID uint, exclude map[uint]struct{}, limit int) ([]model.User, error) {
	logger := config.Logger()
	users, err := s.userRepo.GetActiveUsersByTeam(teamID)
	if err != nil {
		logger.Errorw("failed to list active users", "team_id", teamID, "error", err)
		return nil, err
	}
	logger.Debugw("active team users", "team_id", teamID, "count", len(users))

	filtered := make([]model.User, 0, len(users))
	for _, user := range users {
		if _, skip := exclude[user.ID]; skip {
			continue
		}
		filtered = append(filtered, user)
	}
	logger.Debugw("filtered reviewer candidates", "team_id", teamID, "count", len(filtered), "limit", limit)

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
