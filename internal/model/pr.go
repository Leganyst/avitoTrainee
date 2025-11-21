package model

import "time"

type PullRequest struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	PRID string `gorm:"uniqueIndex; not null"`

	Name     string `gorm:"not null"`
	Status   string `gorm:"not null"`
	AuthorID uint   `gorm:"not null"`

	Author User `gorm:"constraint:OnDelete:CASCADE"`

	/*
		JOIN Table
		pr_reviewers
			pull_request_id (uint)
			user_id (uint)
	*/
	AssignedReviewers []User `gorm:"many2many:pr_reviewers;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt *time.Time
}
