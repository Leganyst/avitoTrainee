package service

import "github.com/Leganyst/avitoTrainee/internal/model"

type PRRepository interface {
	CreatePR(pr *model.PullRequest) error
	GetPRByExternalID(prID string) (*model.PullRequest, error)
	UpdatePR(prID string) error

	AddReviewers(pr *model.PullRequest, reviewers []model.User) error
	ReplaceReviewer(pr *model.PullRequest, oldReviewerID, newReviewerID uint) error

	GetPRsWhereReviewer(userID uint) ([]model.PullRequest, error)
}
