package dto

type UserRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}
