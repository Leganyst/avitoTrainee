package main

import (
	"log"
	"net/http"

	"github.com/Leganyst/avitoTrainee/internal/config"
	"github.com/Leganyst/avitoTrainee/internal/controller/handlers"
	"github.com/Leganyst/avitoTrainee/internal/db"
	"github.com/Leganyst/avitoTrainee/internal/repository"
	"github.com/Leganyst/avitoTrainee/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	log.Println("config loaded")

	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}

	teamRepo := repository.NewTeamRepository(conn)
	userRepo := repository.NewUserRepository(conn)
	prRepo := repository.NewRPRepository(conn)

	teamSvc := service.NewTeamService(teamRepo, userRepo)
	prSvc := service.NewPrService(prRepo, userRepo)
	userSvc := service.NewUserService(userRepo, prRepo)

	r := gin.Default()

	handlers.RegisterRoutes(r, teamSvc, userSvc, prSvc)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("server running on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("server stopped: ", err)
	}
}
