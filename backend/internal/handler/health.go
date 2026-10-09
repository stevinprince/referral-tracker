package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns a simple status response for liveness checks.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
