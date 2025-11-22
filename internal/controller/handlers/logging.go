package handlers

import (
	"strconv"
	"time"

	"github.com/Leganyst/avitoTrainee/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const traceIDKey = "trace_id"

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Request-ID")
		if traceID == "" {
			traceID = strconv.FormatInt(time.Now().UnixNano(), 10)
		}
		c.Set(traceIDKey, traceID)
		c.Writer.Header().Set("X-Request-ID", traceID)

		log := baseLogger().With("traceID", traceID)
		start := time.Now()
		log.Infow("request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		c.Next()

		duration := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		switch {
		case status >= 500:
			log.Errorw("request finished", "status", status, "duration_ms", duration)
		case status >= 400:
			log.Warnw("request finished", "status", status, "duration_ms", duration)
		default:
			log.Infow("request finished", "status", status, "duration_ms", duration)
		}
	}
}

func logger(c *gin.Context) *zap.SugaredLogger {
	traceID, _ := c.Get(traceIDKey)
	if traceID == nil {
		return baseLogger()
	}
	return baseLogger().With("traceID", traceID)
}

func baseLogger() *zap.SugaredLogger {
	if l := config.Logger(); l != nil {
		return l
	}
	return zap.NewNop().Sugar()
}
