package mapper

import (
	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
	"github.com/Leganyst/avitoTrainee/internal/service"
)

// MapAssignmentsByUser переводит сервисные данные в DTO.
func MapAssignmentsByUser(stats []service.AssignmentByUser) []dto.AssignmentByUser {
	items := make([]dto.AssignmentByUser, 0, len(stats))
	for _, s := range stats {
		items = append(items, dto.AssignmentByUser{
			UserID:      s.UserID,
			Username:    s.Username,
			Assignments: s.Assignments,
		})
	}
	return items
}

// MapAssignmentsByPR переводит сервисные данные в DTO.
func MapAssignmentsByPR(stats []service.AssignmentByPR) []dto.AssignmentByPR {
	items := make([]dto.AssignmentByPR, 0, len(stats))
	for _, s := range stats {
		items = append(items, dto.AssignmentByPR{
			PRID:          s.PRID,
			Name:          s.Name,
			ReviewerCount: s.Reviewers,
		})
	}
	return items
}
