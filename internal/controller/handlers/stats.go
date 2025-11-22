package handlers

import (
	"net/http"

	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
	"github.com/Leganyst/avitoTrainee/internal/mapper"
	"github.com/Leganyst/avitoTrainee/internal/service"
	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsSvc service.StatsService
}

func NewStatsHandler(statsSvc service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: statsSvc}
}

func registerStatsRoutes(r gin.IRouter, statsSvc service.StatsService) {
	handler := NewStatsHandler(statsSvc)
	group := r.Group("/stats")

	group.GET("/assignments/by-user", handler.AssignmentsByUser)
	group.GET("/assignments/by-pr", handler.AssignmentsByPR)
}

// AssignmentsByUser godoc
// @Summary      Статистика назначений по пользователям
// @Description  Возвращает список пользователей с количеством назначений на ревью. Пользователи сортируются по количеству назначений по убыванию.
// @Tags         Stats
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.AssignmentByUserResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/stats/assignments/by-user [get]
func (h *StatsHandler) AssignmentsByUser(c *gin.Context) {
	stats, err := h.statsSvc.AssignmentsByUser()
	if err != nil {
		writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
		return
	}

	c.JSON(http.StatusOK, dto.AssignmentByUserResponse{
		Items: mapper.MapAssignmentsByUser(stats),
	})
}

// AssignmentsByPR godoc
// @Summary      Статистика назначений по PR
// @Description  Возвращает список PR с количеством назначенных ревьюверов. Список отсортирован по числу ревьюверов по убыванию.
// @Tags         Stats
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.AssignmentByPRResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/stats/assignments/by-pr [get]
func (h *StatsHandler) AssignmentsByPR(c *gin.Context) {
	stats, err := h.statsSvc.AssignmentsByPR()
	if err != nil {
		writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
		return
	}

	c.JSON(http.StatusOK, dto.AssignmentByPRResponse{
		Items: mapper.MapAssignmentsByPR(stats),
	})
}
