package dto

type PullRequestDTO struct {
	PRID              string   `json:"pull_request_id"`
	Name              string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAT         *string  `json:"created_at,omitempty"`
	MergedAt          *string  `json:"merged_at,omitempty"`
}

type CreatePRResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type MergePRResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type ReassignResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}
