package model

// PRReviewer описывает join-таблицу pr_reviewers с индексами для ускорения вставок/поиска.
type PRReviewer struct {
	PullRequestID uint `gorm:"primaryKey;column:pull_request_id"`
	UserID        uint `gorm:"primaryKey;column:user_id"`
}
