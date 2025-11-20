package dto

// @Description Команда с участниками.
// swagger:model Team
type Team struct {
	// Имя команды.
	TeamName string `json:"team_name" validate:"required" example:"backend"`
	// Участники команды.
	Members []TeamMember `json:"members" validate:"required"`
} // @name Team

// @Description Ответ, содержащий объект team.
// swagger:model TeamResponse
type TeamResponse struct {
	// Объект команды.
	Team Team `json:"team" validate:"required"`
} // @name TeamResponse
