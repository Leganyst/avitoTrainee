package repository

import (
	"github.com/Leganyst/avitoTrainee/internal/config"
	"gorm.io/gorm"
)

type (
	AssignmentStatByUser struct {
		UserID      string
		Username    string
		Assignments int64
	}

	AssignmentStatByPR struct {
		PRID      string
		Name      string
		Reviewers int64
	}

	StatsRepository interface {
		GetAssignmentsByUser() ([]AssignmentStatByUser, error)
		GetAssignmentsByPR() ([]AssignmentStatByPR, error)
	}

	GormStatsRepository struct {
		db *gorm.DB
	}
)

func NewStatsRepository(db *gorm.DB) *GormStatsRepository {
	return &GormStatsRepository{db: db}
}

func (r *GormStatsRepository) GetAssignmentsByUser() ([]AssignmentStatByUser, error) {
	var stats []AssignmentStatByUser
	query := `
		SELECT u.user_id AS user_id, u.username AS username, COUNT(prr.pull_request_id) AS assignments
		FROM pr_reviewers prr
		JOIN users u ON u.id = prr.user_id
		GROUP BY u.id, u.user_id, u.username
		ORDER BY assignments DESC, u.user_id ASC`

	if err := r.db.Raw(query).Scan(&stats).Error; err != nil {
		config.Logger().Errorw("db stats assignments by user failed", "error", err)
		return nil, err
	}

	return stats, nil
}

func (r *GormStatsRepository) GetAssignmentsByPR() ([]AssignmentStatByPR, error) {
	var stats []AssignmentStatByPR
	query := `
		SELECT p.pr_id AS pr_id, p.name AS name, COUNT(prr.user_id) AS reviewers
		FROM pr_reviewers prr
		JOIN pull_requests p ON p.id = prr.pull_request_id
		GROUP BY p.id, p.pr_id, p.name
		ORDER BY reviewers DESC, p.pr_id ASC`

	if err := r.db.Raw(query).Scan(&stats).Error; err != nil {
		config.Logger().Errorw("db stats assignments by pr failed", "error", err)
		return nil, err
	}

	return stats, nil
}
