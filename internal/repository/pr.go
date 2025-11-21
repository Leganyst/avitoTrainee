package repository

import (
	"errors"

	"github.com/Leganyst/avitoTrainee/internal/model"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
	"gorm.io/gorm"
)

type (
	PRRepository interface {
		CreatePR(pr *model.PullRequest) error
		GetPRByExternalID(prID string) (*model.PullRequest, error)
		UpdatePR(pr *model.PullRequest) error

		AddReviewers(pr *model.PullRequest, reviewers []model.User) error
		ReplaceReviewer(pr *model.PullRequest, oldReviewerID uint, newReviewer model.User) error

		GetPRsWhereReviewer(userID uint) ([]model.PullRequest, error)
	}

	GormPRRepository struct {
		db *gorm.DB
	}
)

func NewRPRepository(db *gorm.DB) *GormPRRepository {
	return &GormPRRepository{db}
}

func (r *GormPRRepository) CreatePR(pr *model.PullRequest) error {
	err := r.db.Create(pr).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) { // Для уникальных/дубликатов ключей
			return repoerrs.ErrDuplicate
		}
		return err
	}
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
			return nil, repoerrs.ErrNotFound
		}
		return nil, err
	}
	return &pr, nil
}

func (r *GormPRRepository) UpdatePR(pr *model.PullRequest) error {
	res := r.db.Save(pr)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repoerrs.ErrNotFound
	}
	return nil
}

func (r *GormPRRepository) AddReviewers(pr *model.PullRequest, reviewers []model.User) error {
	reviewersInterface := make([]interface{}, len(reviewers))

	for i, reviewer := range reviewers {
		reviewerCopy := reviewer
		reviewersInterface[i] = &reviewerCopy
	}

	return r.db.Model(pr).Association("AssignedReviewers").Append(reviewersInterface...)
}

func (r *GormPRRepository) ReplaceReviewer(pr *model.PullRequest, oldReviewerID uint, newReviewer model.User) error {
	if err := r.db.
		Model(pr).
		Association("AssignedReviewers").
		Delete(&model.User{ID: oldReviewerID}); err != nil {
		return err
	}

	return r.db.Model(pr).Association("AssignedReviewers").Append(&newReviewer)
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

	return prs, err
}
