package mapper

import (
	"time"

	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
	"github.com/Leganyst/avitoTrainee/internal/model"
)

func MapCreatePRRequestToModel(req dto.CreatePRRequest) model.PullRequest {
	return model.PullRequest{
		PRID:     req.PRID,
		Name:     req.Name,
		AuthorID: req.Author,
	}
}

func MapPullRequestToDTO(pr model.PullRequest) dto.PullRequestDTO {
	return dto.PullRequestDTO{
		PRID:              pr.PRID,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: mapAssignedReviewers(pr.AssignedReviewers),
		CreatedAT:         stringPtrFromTime(pr.CreatedAt),
		MergedAt:          stringPtrFromTimePtr(pr.UpdatedAt),
	}
}

func MapPullRequestShortToDTO(pr model.PullRequest) dto.PullRequestShortDTO {
	return dto.PullRequestShortDTO{
		PRID:     pr.PRID,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
		Status:   pr.Status,
	}
}

func MapPullRequestsToShortDTOs(prs []model.PullRequest) []dto.PullRequestShortDTO {
	shorts := make([]dto.PullRequestShortDTO, 0, len(prs))
	for _, pr := range prs {
		shorts = append(shorts, MapPullRequestShortToDTO(pr))
	}
	return shorts
}

func BuildUserReviewResponse(userID string, prs []model.PullRequest) dto.UserReviewResponse {
	return dto.UserReviewResponse{
		UserID:       userID,
		PullRequests: MapPullRequestsToShortDTOs(prs),
	}
}

func mapAssignedReviewers(reviewers []model.User) []string {
	ids := make([]string, 0, len(reviewers))
	for _, reviewer := range reviewers {
		ids = append(ids, reviewer.UserID)
	}
	return ids
}

func stringPtrFromTime(t time.Time) *string {
	str := t.Format(time.RFC3339)
	return &str
}

func stringPtrFromTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	str := t.Format(time.RFC3339)
	return &str
}
