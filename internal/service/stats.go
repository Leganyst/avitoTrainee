package service

import (
	"github.com/Leganyst/avitoTrainee/internal/repository"
)

type (
	AssignmentByUser struct {
		UserID      string
		Username    string
		Assignments int64
	}

	AssignmentByPR struct {
		PRID      string
		Name      string
		Reviewers int64
	}

	StatsService interface {
		AssignmentsByUser() ([]AssignmentByUser, error)
		AssignmentsByPR() ([]AssignmentByPR, error)
	}

	statsService struct {
		repo repository.StatsRepository
	}
)

func NewStatsService(repo repository.StatsRepository) StatsService {
	return &statsService{repo: repo}
}

func (s *statsService) AssignmentsByUser() ([]AssignmentByUser, error) {
	data, err := s.repo.GetAssignmentsByUser()
	if err != nil {
		return nil, err
	}

	stats := make([]AssignmentByUser, 0, len(data))
	for _, item := range data {
		stats = append(stats, AssignmentByUser{
			UserID:      item.UserID,
			Username:    item.Username,
			Assignments: item.Assignments,
		})
	}
	return stats, nil
}

func (s *statsService) AssignmentsByPR() ([]AssignmentByPR, error) {
	data, err := s.repo.GetAssignmentsByPR()
	if err != nil {
		return nil, err
	}

	stats := make([]AssignmentByPR, 0, len(data))
	for _, item := range data {
		stats = append(stats, AssignmentByPR{
			PRID:      item.PRID,
			Name:      item.Name,
			Reviewers: item.Reviewers,
		})
	}
	return stats, nil
}
