package dto

// @Description Запрос на создание PR.
// swagger:model CreatePRRequest
type CreatePRRequest struct {
	// Идентификатор PR.
	PRID string `json:"pull_request_id" binding:"required" validate:"required" example:"pr-1001"`
	// Название PR.
	Name string `json:"pull_request_name" binding:"required" validate:"required" example:"Add search endpoint"`
	// Автор PR.
	Author string `json:"author_id" binding:"required" validate:"required" example:"u1"`
} // @name CreatePRRequest

// @Description Запрос на merge PR.
// swagger:model MergePRRequest
type MergePRRequest struct {
	// Идентификатор PR.
	PRID string `json:"pull_request_id" binding:"required" validate:"required" example:"pr-1001"`
} // @name MergePRRequest

// @Description Запрос на замену ревьювера.
// swagger:model ReassgnRequest
type ReassgnRequest struct {
	// Идентификатор PR.
	PRID string `json:"pull_request_id" binding:"required" validate:"required" example:"pr-1001"`
	// user_id ревьювера, которого заменяем.
	OldUserID string `json:"old_user_id" binding:"required" validate:"required" example:"u2"`
} // @name ReassgnRequest
