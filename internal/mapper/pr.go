package mapper

import (
	"time"

	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
	"github.com/Leganyst/avitoTrainee/internal/model"
)

// MapCreatePRRequestToModel превращает запрос создания PR в модель.
func MapCreatePRRequestToModel(req dto.CreatePRRequest) model.PullRequest {
	return model.PullRequest{
		PRID:     req.PRID,
		Name:     req.Name,
		AuthorID: req.Author,
	}
}

// MapPullRequestToDTO собирает расширенный DTO из модели PR с ревьюверами.
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

// MapPullRequestShortToDTO делает короткий DTO для списочных ответов.
func MapPullRequestShortToDTO(pr model.PullRequest) dto.PullRequestShortDTO {
	return dto.PullRequestShortDTO{
		PRID:     pr.PRID,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
		Status:   pr.Status,
	}
}

// MapPullRequestsToShortDTOs переводит несколько PR в короткие DTO.
func MapPullRequestsToShortDTOs(prs []model.PullRequest) []dto.PullRequestShortDTO {
	shorts := make([]dto.PullRequestShortDTO, 0, len(prs))
	for _, pr := range prs {
		shorts = append(shorts, MapPullRequestShortToDTO(pr))
	}
	return shorts
}

// BuildUserReviewResponse собирает DTO ответа по ревьюверам.
func BuildUserReviewResponse(userID string, prs []model.PullRequest) dto.UserReviewResponse {
	return dto.UserReviewResponse{
		UserID:       userID,
		PullRequests: MapPullRequestsToShortDTOs(prs),
	}
}

// mapAssignedReviewers вытаскивает user_id из списка ревьюверов.
func mapAssignedReviewers(reviewers []model.User) []string {
	ids := make([]string, 0, len(reviewers))
	for _, reviewer := range reviewers {
		ids = append(ids, reviewer.UserID)
	}
	return ids
}

// stringPtrFromTime нужен, чтобы привести время к RFC3339 и вернуть указатель.
func stringPtrFromTime(t time.Time) *string {
	str := t.Format(time.RFC3339)
	return &str
}

// stringPtrFromTimePtr делает то же самое, но с nullable временем.
func stringPtrFromTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	str := t.Format(time.RFC3339)
	return &str
}
