package dto

// @Description Полное представление PR.
// swagger:model PullRequest
type PullRequest struct {
	// Идентификатор PR.
	PRID string `json:"pull_request_id" validate:"required" example:"pr-1001"`
	// Название PR.
	Name string `json:"pull_request_name" validate:"required" example:"Add search endpoint"`
	// Автор PR.
	AuthorID string `json:"author_id" validate:"required" example:"u1"`
	// Статус PR.
	Status string `json:"status" validate:"required" example:"OPEN"`
	// Назначенные ревьюверы (до двух user_id).
	AssignedReviewers []string `json:"assigned_reviewers" validate:"required" example:"u2,u3"`
	// Время создания.
	CreatedAT *string `json:"created_at,omitempty" example:"2025-10-25T12:00:00Z"`
	// Время merge (если есть).
	MergedAt *string `json:"merged_at,omitempty" example:"2025-10-26T09:30:00Z"`
} // @name PullRequest

// @Description Ответ на создание PR.
// swagger:model CreatePRResponse
type CreatePRResponse struct {
	// Созданный PR.
	PR PullRequest `json:"pr" validate:"required"`
} // @name CreatePRResponse

// @Description Ответ на merge PR.
// swagger:model MergePRResponse
type MergePRResponse struct {
	// PR после merge.
	PR PullRequest `json:"pr" validate:"required"`
} // @name MergePRResponse

// @Description Ответ на замену ревьювера.
// swagger:model ReassignResponse
type ReassignResponse struct {
	// Текущий PR.
	PR PullRequest `json:"pr" validate:"required"`
	// user_id, который заменил предыдущего ревьювера.
	ReplacedBy string `json:"replaced_by" validate:"required" example:"u5"`
} // @name ReassignResponse
