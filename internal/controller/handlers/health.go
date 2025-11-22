package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary      Health Check
// @Description  Returns service health status.
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /healthcheck [get]
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "UP",
		"timestamp": time.Now().UTC(),
	})
}
