package handlers

import (
	"github.com/Leganyst/avitoTrainee/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes прокидывает зависимости в хэндлеры и вешает эндпоинты.
func RegisterRoutes(r *gin.Engine,
	teamSvc service.TeamService,
	userSvc service.UserService,
	prSvc service.PRService,
	statsSvc service.StatsService,
) {
	r.Use(requestLoggerMiddleware())

	r.GET("/healthcheck", healthCheckHandler)

	api := r.Group("/api")

	registerTeamRoutes(api, teamSvc)
	registerUserRoutes(api, userSvc)
	registerPRRoutes(api, prSvc)
	registerStatsRoutes(api, statsSvc)
}
