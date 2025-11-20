package mapper

import (
	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
	"github.com/Leganyst/avitoTrainee/internal/model"
)

// MapCreateTeamRequestToModel переводит входящий запрос создания команды в модель.
func MapCreateTeamRequestToModel(req dto.CreateTeamRequest) model.Team {
	return model.Team{
		Name:  req.TeamName,
		Users: MapTeamMemberDTOsToUsers(req.Members),
	}
}

// MapUsersToTeamMemberDTO переводит модельных юзеров в DTO участников команды.
func MapUsersToTeamMemberDTO(users []model.User) []dto.TeamMember {
	members := make([]dto.TeamMember, 0, len(users))
	for _, user := range users {
		members = append(members, dto.TeamMember{
			UserID:   user.UserID,
			Username: user.Username,
			IsActive: user.IsActive,
		})
	}
	return members
}
