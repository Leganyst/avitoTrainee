package handlers

import (
	"net/http"

	"github.com/Leganyst/avitoTrainee/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes прокидывает зависимости в хэндлеры и вешает эндпоинты.
func RegisterRoutes(r *gin.Engine,
	teamSvc service.TeamService,
	userSvc service.UserService,
	prSvc service.PRService,
) {
	r.Use(requestLoggerMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	api := r.Group("/api")

	registerTeamRoutes(api, teamSvc)
	registerUserRoutes(api, userSvc)
	registerPRRoutes(api, prSvc)
}
