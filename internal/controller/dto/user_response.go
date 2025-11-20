package dto

// @Description Представление пользователя.
// swagger:model User
type User struct {
	// Идентификатор пользователя.
	UserID string `json:"user_id" validate:"required" example:"u2"`
	// Имя пользователя.
	Username string `json:"username" validate:"required" example:"Bob"`
	// Название команды.
	TeamName string `json:"team_name" validate:"required" example:"backend"`
	// Флаг активности.
	IsActive bool `json:"is_active" validate:"required" example:"true"`
} // @name User

// @Description Ответ с пользователем.
// swagger:model UserResponse
type UserResponse struct {
	// Пользователь.
	User User `json:"user" validate:"required"`
} // @name UserResponse
