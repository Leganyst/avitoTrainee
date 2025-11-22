package dto

// @Description Количество назначений по пользователям.
// swagger:model AssignmentByUser
type AssignmentByUser struct {
	// user_id ревьювера.
	UserID string `json:"user_id" example:"u1"`
	// username ревьювера.
	Username string `json:"username" example:"Alice"`
	// Сколько раз этот пользователь назначался ревьювером.
	Assignments int64 `json:"assignments" example:"3"`
} // @name AssignmentByUser

// @Description Количество назначений по PR.
// swagger:model AssignmentByPR
type AssignmentByPR struct {
	// Идентификатор PR.
	PRID string `json:"pull_request_id" example:"pr-1001"`
	// Название PR.
	Name string `json:"pull_request_name" example:"Add search endpoint"`
	// Число назначенных ревьюверов для PR.
	ReviewerCount int64 `json:"reviewer_count" example:"2"`
} // @name AssignmentByPR

// @Description Ответ со списком назначений по пользователям.
// swagger:model AssignmentByUserResponse
type AssignmentByUserResponse struct {
	Items []AssignmentByUser `json:"items"`
} // @name AssignmentByUserResponse

// @Description Ответ со списком назначений по PR.
// swagger:model AssignmentByPRResponse
type AssignmentByPRResponse struct {
	Items []AssignmentByPR `json:"items"`
} // @name AssignmentByPRResponse
