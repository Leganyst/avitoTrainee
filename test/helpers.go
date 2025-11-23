package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Leganyst/avitoTrainee/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func connectTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "55432")
	user := getenv("DB_USER", "app_test")
	pass := getenv("DB_PASS", "app_test")
	name := getenv("DB_NAME", "app_test")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, name)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info),
		TranslateError: true,
	})
	if err != nil {
		t.Fatalf("failed to connect test DB: %v", err)
	}
	return db
}

func migrateTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(
		&model.Team{},
		&model.User{},
		&model.PullRequest{},
		&model.PRReviewer{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
}

func getenv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func prepareDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	_ = db.Migrator().DropTable("pr_reviewers")

	err := db.Exec(
		"TRUNCATE TABLE pull_requests, users, teams RESTART IDENTITY CASCADE",
	).Error

	if err != nil {
		// Ошибка "relation does not exist" → таблиц нет → просто мигрируем
		if strings.Contains(err.Error(), "42P01") || strings.Contains(err.Error(), "does not exist") {
			t.Log("No tables to truncate — running migrations only")
			migrateTestDB(t, db)
			return
		}

		t.Fatalf("unexpected reset error: %v", err)
	}

	migrateTestDB(t, db)
}
