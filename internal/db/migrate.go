package db

import (
	"github.com/Leganyst/avitoTrainee/internal/model"
	"gorm.io/gorm"
)

func Migrate(conn *gorm.DB) error {
	// if err := conn.SetupJoinTable(&model.PullRequest{}, "AssignedReviewers", &model.User{}); err != nil {
	// 	return err
	// }
	return conn.AutoMigrate(&model.Team{}, &model.User{}, &model.PullRequest{})
}
