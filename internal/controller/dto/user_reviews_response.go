package dto

// @Description Короткое описание PR.
// swagger:model PullRequestShort
type PullRequestShort struct {
	// Идентификатор PR.
	PRID string `json:"pull_request_id" validate:"required" example:"pr-1001"`
	// Название PR.
	Name string `json:"pull_request_name" validate:"required" example:"Add search endpoint"`
	// Автор PR.
	AuthorID string `json:"author_id" validate:"required" example:"u1"`
	// Статус PR.
	Status string `json:"status" validate:"required" example:"OPEN"`
} // @name PullRequestShort

// @Description PR, где пользователь ревьювер.
// swagger:model UserReviewResponse
type UserReviewResponse struct {
	// Идентификатор пользователя.
	UserID string `json:"user_id" validate:"required" example:"u2"`
	// Список PR.
	PullRequests []PullRequestShort `json:"pull_requests" validate:"required"`
} // @name UserReviewResponse
