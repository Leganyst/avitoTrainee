package controller

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	api := r.Group("/api")

	registerUserRoutes(api, db)
	registerTeamRoutes(api, db)
	registerPRRoutes(api, db)

}
