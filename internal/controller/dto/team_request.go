package dto

// @Description Участник команды.
// swagger:model TeamMember
type TeamMember struct {
	// user_id участника.
	UserID string `json:"user_id" binding:"required" validate:"required" example:"u1"`
	// username участника.
	Username string `json:"username" binding:"required" validate:"required" example:"Alice"`
	// Признак активности.
	IsActive bool `json:"is_active" binding:"required" validate:"required" example:"true"`
} // @name TeamMember

// @Description Запрос на создание команды.
// swagger:model CreateTeamRequest
type CreateTeamRequest struct {
	// Имя команды.
	TeamName string `json:"team_name" binding:"required" validate:"required" example:"backend"`
	// Массив участников команды.
	Members []TeamMember `json:"members" binding:"required,dive" validate:"required,dive"`
} // @name CreateTeamRequest
