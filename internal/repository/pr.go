package repository

import (
	"github.com/Leganyst/avitoTrainee/internal/model"
	"gorm.io/gorm"
)

type GormPRRepository struct {
	db *gorm.DB
}

func NewRPRepository(db *gorm.DB) *GormPRRepository {
	return &GormPRRepository{db}
}

func (r *GormPRRepository) UpdatePR(pr *model.PullRequest) error {
	return r.db.Save(pr).Error
}

func (r *GormPRRepository) AddReviewers(pr *model.PullRequest, reviewers []model.User) error {
	reviewersInterface := make([]interface{}, len(reviewers))
	for i, reviewer := range reviewers {
		reviewersInterface[i] = reviewer
	}
	return r.db.Association("AssignedReviewers").Append(reviewersInterface...)
}

func (r *GormPRRepository) ReplaceReviewer(pr *model.PullRequest, oldReviewerID, newReviewerID uint) error {
	if err := r.db.Model(pr).Association("AssignedReviewers").
		Delete(&model.User{ID: oldReviewerID}); err != nil {
		return err
	}

	return r.db.Model(pr).Association("AssignedReviewers").
		Append(&model.User{ID: newReviewerID})
}

func (r *GormPRRepository) GetPRsWhereReviewer(userID uint) ([]model.PullRequest, error) {
	var prs []model.PullRequest
	err := r.db.
		Model(&model.PullRequest{}).
		Joins("JOIN pr_reviewers ON pr_reviewers.pull_request_id = pull_requests.id").
		Where("pr_reviewers.user_id = ?", userID).
		Preload("AssignedReviewers").
		Find(&prs).Error

	return prs, err
}
