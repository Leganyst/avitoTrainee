package mapper

import (
	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
	"github.com/Leganyst/avitoTrainee/internal/model"
)

// MapUserToDTO превращает модель User в DTO для ответов.
func MapUserToDTO(user model.User) dto.UserDTO {
	return dto.UserDTO{
		UserID:   user.UserID,
		Username: user.Username,
		TeamName: user.Team.Name,
		IsActive: user.IsActive,
	}
}

// MapUsersToDTO превращает список User в DTO без лишней логики.
func MapUsersToDTO(users []model.User) []dto.UserDTO {
	dtos := make([]dto.UserDTO, 0, len(users))
	for _, user := range users {
		dtos = append(dtos, MapUserToDTO(user))
	}
	return dtos
}

// MapTeamMemberDTOToUser собирает модель User из DTO участника команды.
func MapTeamMemberDTOToUser(member dto.TeamMemberDTO) model.User {
	return model.User{
		UserID:   member.UserID,
		Username: member.Username,
		IsActive: member.IsActive,
	}
}

// MapTeamMemberDTOsToUsers собирает модели User из списка DTO участников.
func MapTeamMemberDTOsToUsers(members []dto.TeamMemberDTO) []model.User {
	users := make([]model.User, len(members))
	for i := range members {
		users[i] = MapTeamMemberDTOToUser(members[i])
	}
	return users
}

// MapUserRequestToUser конвертирует payload запроса в модель User.
func MapUserRequestToUser(req dto.UserRequest) model.User {
	return model.User{
		UserID:   req.UserID,
		IsActive: req.IsActive,
	}
}
