package repository

import (
	"errors"
	"strings"

	"github.com/Leganyst/avitoTrainee/internal/config"
	"github.com/Leganyst/avitoTrainee/internal/model"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	PRRepository interface {
		CreatePR(pr *model.PullRequest) error
		GetPRByExternalID(prID string) (*model.PullRequest, error)
		UpdatePR(pr *model.PullRequest) error

		AddReviewers(pr *model.PullRequest, reviewers []model.User) error
		ReplaceReviewer(pr *model.PullRequest, oldReviewerID uint, newReviewer model.User) error
		ReplaceReviewers(prID uint, reviewerIDs []uint) error

		GetPRsWhereReviewer(userID uint) ([]model.PullRequest, error)
		GetOpenPRsByReviewerIDs(reviewerIDs []uint) ([]model.PullRequest, error)
	}

	GormPRRepository struct {
		db *gorm.DB
	}
)

func NewPRRepository(db *gorm.DB) *GormPRRepository {
	return &GormPRRepository{db}
}

func (r *GormPRRepository) CreatePR(pr *model.PullRequest) error {
	if err := r.db.Create(pr).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || isUniqueViolation(err) {
			config.Logger().Warnw("db PR duplicate", "pr_id", pr.PRID)
			return repoerrs.ErrDuplicate
		}
		config.Logger().Errorw("db create PR failed", "pr_id", pr.PRID, "error", err)
		return err
	}
	config.Logger().Debugw("db PR created", "pr_id", pr.PRID, "author_id", pr.AuthorID)
	return nil
}

func (r *GormPRRepository) GetPRByExternalID(prID string) (*model.PullRequest, error) {
	var pr model.PullRequest
	if err := r.db.
		Preload("Author").
		Preload("AssignedReviewers").
		Where("pr_id = ?", prID).
		First(&pr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			config.Logger().Warnw("db PR not found", "pr_id", prID)
			return nil, repoerrs.ErrNotFound
		}
		config.Logger().Errorw("db get PR failed", "pr_id", prID, "error", err)
		return nil, err
	}
	config.Logger().Debugw("db PR loaded", "pr_id", prID, "reviewers", len(pr.AssignedReviewers))
	return &pr, nil
}

func (r *GormPRRepository) UpdatePR(pr *model.PullRequest) error {
	res := r.db.Save(pr)
	if res.Error != nil {
		config.Logger().Errorw("db update PR failed", "pr_id", pr.PRID, "error", res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		config.Logger().Warnw("db update PR no rows", "pr_id", pr.PRID)
		return repoerrs.ErrNotFound
	}
	config.Logger().Debugw("db PR updated", "pr_id", pr.PRID)
	return nil
}

func (r *GormPRRepository) AddReviewers(pr *model.PullRequest, reviewers []model.User) error {
	if len(reviewers) == 0 {
		return nil
	}

	rows := make([]map[string]interface{}, 0, len(reviewers))
	for _, reviewer := range reviewers {
		rows = append(rows, map[string]interface{}{
			"pull_request_id": pr.ID,
			"user_id":         reviewer.ID,
		})
	}

	if err := r.db.Table("pr_reviewers").Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
		config.Logger().Errorw("db append reviewers failed", "pr_id", pr.PRID, "error", err)
		return err
	}
	config.Logger().Debugw("db reviewers appended", "pr_id", pr.PRID, "count", len(reviewers))
	return nil
}

func (r *GormPRRepository) ReplaceReviewer(pr *model.PullRequest, oldReviewerID uint, newReviewer model.User) error {
	if err := r.db.
		Model(pr).
		Association("AssignedReviewers").
		Delete(&model.User{ID: oldReviewerID}); err != nil {
		config.Logger().Errorw("db delete reviewer failed", "pr_id", pr.PRID, "old_user", oldReviewerID, "error", err)
		return err
	}

	if err := r.db.Model(pr).Association("AssignedReviewers").Append(&newReviewer); err != nil {
		config.Logger().Errorw("db append new reviewer failed", "pr_id", pr.PRID, "new_user", newReviewer.UserID, "error", err)
		return err
	}
	config.Logger().Debugw("db reviewer replaced", "pr_id", pr.PRID, "old_user", oldReviewerID, "new_user", newReviewer.UserID)
	return nil
}

// ReplaceReviewers заменяет весь список ревьюверов за один проход.
func (r *GormPRRepository) ReplaceReviewers(prID uint, reviewerIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pull_request_id = ?", prID).Delete(&model.PRReviewer{}).Error; err != nil {
			return err
		}
		if len(reviewerIDs) == 0 {
			return nil
		}

		rows := make([]map[string]interface{}, 0, len(reviewerIDs))
		for _, id := range reviewerIDs {
			rows = append(rows, map[string]interface{}{
				"pull_request_id": prID,
				"user_id":         id,
			})
		}

		if err := tx.Table("pr_reviewers").Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *GormPRRepository) GetPRsWhereReviewer(userID uint) ([]model.PullRequest, error) {
	var prs []model.PullRequest
	err := r.db.
		Model(&model.PullRequest{}).
		Joins("JOIN pr_reviewers ON pr_reviewers.pull_request_id = pull_requests.id").
		Where("pr_reviewers.user_id = ?", userID).
		Preload("Author").
		Preload("AssignedReviewers").
		Find(&prs).Error

	if err != nil {
		config.Logger().Errorw("db list PRs for reviewer failed", "user_id", userID, "error", err)
		return nil, err
	}
	config.Logger().Debugw("db PRs for reviewer loaded", "user_id", userID, "count", len(prs))
	return prs, err
}

func (r *GormPRRepository) GetOpenPRsByReviewerIDs(reviewerIDs []uint) ([]model.PullRequest, error) {
	if len(reviewerIDs) == 0 {
		return nil, nil
	}

	var prs []model.PullRequest
	err := r.db.
		Model(&model.PullRequest{}).
		Distinct("pull_requests.id").
		Joins("JOIN pr_reviewers ON pr_reviewers.pull_request_id = pull_requests.id").
		Where("pr_reviewers.user_id IN ?", reviewerIDs).
		Where("pull_requests.status = ?", "OPEN").
		Preload("Author").
		Preload("AssignedReviewers").
		Find(&prs).Error
	if err != nil {
		config.Logger().Errorw("db open PRs by reviewer ids failed", "reviewer_ids", reviewerIDs, "error", err)
		return nil, err
	}
	config.Logger().Debugw("db open PRs by reviewer ids loaded", "reviewer_ids_len", len(reviewerIDs), "prs", len(prs))
	return prs, nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key value")
}
