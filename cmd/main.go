package main

import (
	"net/http"

	docs "github.com/Leganyst/avitoTrainee/docs"
	"github.com/Leganyst/avitoTrainee/internal/config"
	"github.com/Leganyst/avitoTrainee/internal/controller/handlers"
	"github.com/Leganyst/avitoTrainee/internal/db"
	"github.com/Leganyst/avitoTrainee/internal/repository"
	"github.com/Leganyst/avitoTrainee/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           PR Reviewer Service API
// @version         1.0
// @description     Service for assigning reviewers to pull requests.
// @BasePath        /
func main() {
	cfg := config.Load()
	if err := config.InitLogger(cfg.LogLevel); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer config.Logger().Sync()
	config.Logger().Infow("config loaded", "port", cfg.Port, "logLevel", cfg.LogLevel)

	docs.SwaggerInfo.BasePath = "/"

	conn, err := db.Connect(cfg)
	if err != nil {
		config.Logger().Fatalw("cannot connect to database", "error", err)
	}

	if err := db.Migrate(conn); err != nil {
		config.Logger().Fatalw("auto-migrate failed", "error", err)
	}

	teamRepo := repository.NewTeamRepository(conn)
	userRepo := repository.NewUserRepository(conn)
	prRepo := repository.NewPRRepository(conn)

	teamSvc := service.NewTeamService(teamRepo, userRepo)
	prSvc := service.NewPrService(prRepo, userRepo)
	userSvc := service.NewUserService(userRepo, prRepo)

	r := gin.Default()

	handlers.RegisterRoutes(r, teamSvc, userSvc, prSvc)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	config.Logger().Info("server running on :8080")
	if err := srv.ListenAndServe(); err != nil {
		config.Logger().Fatalw("server stopped", "error", err)
	}
}
