package main

import (
	"log"

	"github.com/Leganyst/avitoTrainee/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	log.Println("config loaded")

	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}

	r := gin.Default()

}
