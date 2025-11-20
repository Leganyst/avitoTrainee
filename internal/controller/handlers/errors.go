package handlers

import "github.com/gin-gonic/gin"

const (
	errorCodeBadRequest  = "BAD_REQUEST"
	errorCodeInternal    = "INTERNAL"
	errorCodeNotFound    = "NOT_FOUND"
	errorCodeTeamExists  = "TEAM_EXISTS"
	errorCodePRExists    = "PR_EXISTS"
	errorCodePRMerged    = "PR_MERGED"
	errorCodeNotAssigned = "NOT_ASSIGNED"
	errorCodeNoCandidate = "NO_CANDIDATE"
)

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
