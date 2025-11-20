package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerUserRoutes(r *gin.Engine, db *gorm.DB) {

	userRouters = r.Group("/user")

}
