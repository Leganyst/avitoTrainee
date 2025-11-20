package handlers

import (
	"github.com/Leganyst/avitoTrainee/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB,
					teamSvc service.TeamService,
					userSvc service.UserService,
					prSvc service.PRService) {

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	teamHandler := NewTeamHandler(teamSvc)
	userHandler := 

	api := r.Group("/api")

	{
		api.POST("/team/add", teamHandler.CreateTeam)
		api.GET("/team/get", teanHandler.GetTeam)
	}

}
