package dto

type CreatePRRequest struct {
	PRID   string `json:"pull_request_id"`
	Name   string `json:"pull_request_name"`
	Author string `json:"author_id"`
}

type MergePRRequest struct {
	PRID string `json:"pull_request_id"`
}

type ReassgnRequest struct {
	PRID      string `json:"pull_request_id"`
	OldUserID string `json:"old_user_id"`
}
