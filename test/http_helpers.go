package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Leganyst/avitoTrainee/internal/controller/handlers"
	"github.com/Leganyst/avitoTrainee/internal/repository"
	"github.com/Leganyst/avitoTrainee/internal/service"
	"github.com/gin-gonic/gin"
)

type apiTestServer struct {
	router *gin.Engine
}

func newAPITestServer(t *testing.T) *apiTestServer {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db := connectTestDB(t)
	prepareDB(t, db)

	if sqlDB, err := db.DB(); err == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	teamSvc := service.NewTeamService(teamRepo, userRepo)
	userSvc := service.NewUserService(userRepo, prRepo, teamRepo)
	prSvc := service.NewPrService(prRepo, userRepo)
	statsSvc := service.NewStatsService(statsRepo)

	router := gin.New()
	router.Use(gin.Recovery())
	handlers.RegisterRoutes(router, teamSvc, userSvc, prSvc, statsSvc)

	return &apiTestServer{router: router}
}

func (s *apiTestServer) doRequest(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}
