package dto

// @Description Запрос на смену активности пользователя.
// swagger:model UserRequest
type UserRequest struct {
	// Идентификатор пользователя.
	UserID string `json:"user_id" binding:"required" validate:"required" example:"u2"`
	// Значение флага активности.
	IsActive bool `json:"is_active" binding:"required" validate:"required" example:"false"`
} // @name UserRequest
