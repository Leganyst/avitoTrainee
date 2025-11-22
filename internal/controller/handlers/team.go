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

// CreateTeam godoc
// @Summary      Создать команду
// @Description  Создаёт команду и пользователей, если их ещё нет.
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateTeamRequest  true  "Данные команды"
// @Success      201      {object}  dto.TeamResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/team/add [post]
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	log := logger(c)
	var req dto.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnw("invalid create team payload", "error", err)
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}
	if req.TeamName == "" {
		log.Warnw("missing team_name in create team request")
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "team_name is required")
		return
	}
	log.Debugw("create team payload", "request", req)

	members := mapper.MapTeamMemberDTOsToUsers(req.Members)
	team, err := h.teamSvc.CreateTeam(req.TeamName, members)
	if err != nil {
		switch {
		case errors.Is(err, serviceerrs.ErrTeamExists):
			log.Warnw("team already exists", "team_name", req.TeamName)
			writeError(c, http.StatusBadRequest, errorCodeTeamExists, err.Error())
		default:
			log.Errorw("failed to create team", "team_name", req.TeamName, "error", err)
			writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
		}
		return
	}

	teamDTO := dto.Team{
		TeamName: team.Name,
		Members:  mapper.MapUsersToTeamMemberDTO(team.Users),
	}

	c.JSON(http.StatusCreated, dto.TeamResponse{
		Team: teamDTO,
	})
	log.Infow("team created", "team_name", team.Name, "members", len(team.Users))
}

// GetTeam godoc
// @Summary      Получить команду
// @Description  Возвращает команду и участников по имени.
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Param        team_name  query     string  true  "Имя команды"
// @Success      200        {object}  dto.Team
// @Failure      400        {object}  dto.ErrorResponse
// @Failure      404        {object}  dto.ErrorResponse
// @Failure      500        {object}  dto.ErrorResponse
// @Router       /api/team/get [get]
func (h *TeamHandler) GetTeam(c *gin.Context) {
	log := logger(c)
	teamName := c.Query("team_name")
	if teamName == "" {
		log.Warnw("team_name query parameter missing")
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "team_name is required")
		return
	}
	log.Debugw("get team request", "team_name", teamName)

	team, err := h.teamSvc.GetTeam(teamName)
	if err != nil {
		switch {
		case errors.Is(err, serviceerrs.ErrTeamNotFound):
			log.Warnw("team not found", "team_name", teamName)
			writeError(c, http.StatusNotFound, errorCodeNotFound, err.Error())
		default:
			log.Errorw("failed to get team", "team_name", teamName, "error", err)
			writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
		}
		return
	}

	c.JSON(http.StatusOK, dto.Team{
		TeamName: team.Name,
		Members:  mapper.MapUsersToTeamMemberDTO(team.Users),
	})
	log.Infow("team fetched", "team_name", team.Name, "members", len(team.Users))
}
