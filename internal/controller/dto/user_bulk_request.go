package dto

// @Description Запрос на массовую деактивацию пользователей команды.
// swagger:model BulkDeactivateRequest
type BulkDeactivateRequest struct {
	// Имя команды, в которой отключаются пользователи.
	TeamName string `json:"team_name" binding:"required" example:"backend"`
	// user_id пользователей для деактивации.
	UserIDs []string `json:"user_ids" binding:"required" example:"u1,u2,u3"`
} // @name BulkDeactivateRequest

// @Description Ответ на массовую деактивацию.
// swagger:model BulkDeactivateResponse
type BulkDeactivateResponse struct {
	// Имя команды.
	Team string `json:"team"`
	// Сколько пользователей деактивировано.
	Deactivated int `json:"deactivated"`
	// Сколько замен ревьюверов выполнено.
	Reassigned int `json:"reassigned"`
	// Сколько замен пропущено из-за отсутствия кандидатов.
	Skipped int `json:"skipped"`
	// Сколько открытых PR затронуто.
	AffectedPRs int `json:"affected_prs"`
} // @name BulkDeactivateResponse
