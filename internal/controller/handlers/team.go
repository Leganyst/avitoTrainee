package handlers

import (
	"errors"
	"net/http"

	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
	"github.com/Leganyst/avitoTrainee/internal/mapper"
	"github.com/Leganyst/avitoTrainee/internal/service"
	serviceerrs "github.com/Leganyst/avitoTrainee/internal/service/errs"
	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamSvc service.TeamService
}

func NewTeamHandler(teamSvc service.TeamService) *TeamHandler {
	return &TeamHandler{teamSvc: teamSvc}
}

func registerTeamRoutes(r gin.IRouter, teamSvc service.TeamService) {
	handler := NewTeamHandler(teamSvc)

	group := r.Group("/team")
	group.POST("/add", handler.CreateTeam)
	group.GET("/get", handler.GetTeam)
}

// CreateTeam создаёт команду и пользователей
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req dto.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}
	if req.TeamName == "" {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "team_name is required")
		return
	}

	members := mapper.MapTeamMemberDTOsToUsers(req.Members)
	team, err := h.teamSvc.CreateTeam(req.TeamName, members)
	if err != nil {
		switch {
		case errors.Is(err, serviceerrs.ErrTeamExists):
			writeError(c, http.StatusBadRequest, errorCodeTeamExists, err.Error())
		default:
			writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
		}
		return
	}

	c.JSON(http.StatusCreated, dto.TeamResponse{
		TeamName: team.Name,
		Members:  mapper.MapUsersToTeamMemberDTO(team.Users),
	})
}

// GetTeam возвращает команду с участниками по имени.
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "team_name is required")
		return
	}

	team, err := h.teamSvc.GetTeam(teamName)
	if err != nil {
		switch {
		case errors.Is(err, serviceerrs.ErrTeamNotFound):
			writeError(c, http.StatusNotFound, errorCodeNotFound, err.Error())
		default:
			writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
		}
		return
	}

	c.JSON(http.StatusOK, dto.TeamResponse{
		TeamName: team.Name,
		Members:  mapper.MapUsersToTeamMemberDTO(team.Users),
	})
}
