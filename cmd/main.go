package main

import (
	"log"
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
	log.Println("config loaded")

	docs.SwaggerInfo.BasePath = "/"

	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}

	if err := db.Migrate(conn); err != nil {
		log.Fatal("auto-migrate failed: ", err)
	}

	teamRepo := repository.NewTeamRepository(conn)
	userRepo := repository.NewUserRepository(conn)
	prRepo := repository.NewRPRepository(conn)

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

	log.Println("server running on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("server stopped: ", err)
	}
}
