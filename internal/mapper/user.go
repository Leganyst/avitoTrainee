package mapper

import (
	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
	"github.com/Leganyst/avitoTrainee/internal/model"
)

func MapUserToDTO(user model.User) dto.UserDTO {
	return dto.UserDTO{
		UserID:   user.UserID,
		Username: user.Username,
		TeamName: user.Team.Name,
		IsActive: user.IsActive,
	}
}

func MapUsersToDTO(users []model.User) []dto.UserDTO {
	dtos := make([]dto.UserDTO, 0, len(users))
	for _, user := range users {
		dtos = append(dtos, MapUserToDTO(user))
	}
	return dtos
}

func MapTeamMemberDTOToUser(member dto.TeamMemberDTO) model.User {
	return model.User{
		UserID:   member.UserID,
		Username: member.Username,
		IsActive: member.IsActive,
	}
}

func MapTeamMemberDTOsToUsers(members []dto.TeamMemberDTO) []model.User {
	users := make([]model.User, len(members))
	for i := range members {
		users[i] = MapTeamMemberDTOToUser(members[i])
	}
	return users
}

func MapUserRequestToUser(req dto.UserRequest) model.User {
	return model.User{
		UserID:   req.UserID,
		IsActive: req.IsActive,
	}
}
