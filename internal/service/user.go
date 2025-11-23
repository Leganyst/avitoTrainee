package service

import (
	"errors"
	"fmt"

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
		BulkDeactivate(teamName string, userIDs []string) (*BulkDeactivateResult, error)
	}

	userService struct {
		userRepo repository.UserRepository
		prRepo   repository.PRRepository
		teamRepo repository.TeamRepository
	}

	BulkDeactivateResult struct {
		TeamName             string
		DeactivatedUsers     int
		ReassignmentsDone    int
		ReassignmentsSkipped int
		AffectedPullRequests int
	}
)

func NewUserService(userRepo repository.UserRepository, prRepo repository.PRRepository, teamRepo repository.TeamRepository) UserService {
	return &userService{
		userRepo: userRepo,
		prRepo:   prRepo,
		teamRepo: teamRepo,
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

// BulkDeactivate деактивирует пользователей команды и безопасно переназначает их в открытых PR.
// Операция накладная, как минимум O(n * m), где N - количество PR, M - пользователей которых придется переназначать
func (s *userService) BulkDeactivate(teamName string, userIDs []string) (*BulkDeactivateResult, error) {
	logger := config.Logger()
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("user_ids is required")
	}

	team, err := s.teamRepo.GetTeamByName(teamName)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("bulk deactivate team not found", "team_name", teamName)
			return nil, serviceerrs.ErrTeamNotFound
		}
		logger.Errorw("bulk deactivate get team failed", "team_name", teamName, "error", err)
		return nil, err
	}

	toDeactivate, err := s.userRepo.BulkDeactivate(team.ID, userIDs)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			logger.Warnw("bulk deactivate users not found", "team_name", teamName, "user_ids", userIDs)
			return nil, serviceerrs.ErrUserNotFound
		}
		logger.Errorw("bulk deactivate failed", "team_name", teamName, "error", err)
		return nil, err
	}

	deactivatedByID := make(map[uint]model.User, len(toDeactivate))
	reviewerIDs := make([]uint, 0, len(toDeactivate))
	for _, u := range toDeactivate {
		deactivatedByID[u.ID] = u
		reviewerIDs = append(reviewerIDs, u.ID)
	}

	prs, err := s.prRepo.GetOpenPRsByReviewerIDs(reviewerIDs)
	if err != nil {
		logger.Errorw("bulk deactivate fetch open prs failed", "team_name", teamName, "error", err)
		return nil, err
	}

	result := &BulkDeactivateResult{
		TeamName:         teamName,
		DeactivatedUsers: len(toDeactivate),
	}

	for i := range prs {
		pr := &prs[i]
		affected := false

		for idx, reviewer := range pr.AssignedReviewers {
			if _, shouldReplace := deactivatedByID[reviewer.ID]; !shouldReplace {
				continue
			}

			excluded := make(map[uint]struct{}, len(pr.AssignedReviewers)+2)
			excluded[pr.AuthorID] = struct{}{}
			for _, r := range pr.AssignedReviewers {
				excluded[r.ID] = struct{}{}
			}

			candidate, err := s.selectReplacementCandidate(reviewer.TeamID, excluded)
			if err != nil {
				logger.Errorw("bulk deactivate select replacement failed", "pr_id", pr.PRID, "old_user", reviewer.UserID, "error", err)
				return nil, err
			}
			if candidate == nil {
				result.ReassignmentsSkipped++
				continue
			}

			if err := s.prRepo.ReplaceReviewer(pr, reviewer.ID, *candidate); err != nil {
				logger.Errorw("bulk replace reviewer failed", "pr_id", pr.PRID, "old_user", reviewer.UserID, "new_user", candidate.UserID, "error", err)
				return nil, err
			}

			pr.AssignedReviewers[idx] = *candidate
			result.ReassignmentsDone++
			affected = true
		}

		if affected {
			result.AffectedPullRequests++
		}
	}

	logger.Infow("bulk deactivate completed", "team_name", teamName, "deactivated", result.DeactivatedUsers, "reassigned", result.ReassignmentsDone, "skipped", result.ReassignmentsSkipped, "prs", result.AffectedPullRequests)
	return result, nil
}

// selectReplacementCandidate выбирает активного пользователя команды с учётом исключений.
func (s *userService) selectReplacementCandidate(teamID uint, excluded map[uint]struct{}) (*model.User, error) {
	users, err := s.userRepo.GetActiveUsersByTeam(teamID)
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if _, skip := excluded[u.ID]; skip {
			continue
		}
		return &u, nil
	}
	return nil, nil
}
