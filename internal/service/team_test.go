package service

import (
	"errors"
	"testing"

	"github.com/Leganyst/avitoTrainee/internal/model"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
	serviceerrs "github.com/Leganyst/avitoTrainee/internal/service/errs"
)

func TestTeamService_CreateTeam_Success(t *testing.T) {
	teamRepo := &stubTeamRepo{}
	userRepo := &stubUserRepo{}
	svc := teamService{teamRepo: teamRepo, userRepo: userRepo}

	members := []model.User{
		{UserID: "u1", Username: "Alice"},
		{UserID: "u2", Username: "Bob"},
	}

	team, err := svc.CreateTeam("backend", members)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team.ID != 42 {
		t.Fatalf("expected team ID assigned to 42, got %d", team.ID)
	}
	if len(team.Users) != len(members) {
		t.Fatalf("expected %d users in team, got %d", len(members), len(team.Users))
	}
	for _, user := range userRepo.created {
		if user.TeamID != 42 {
			t.Fatalf("expected user TeamID to be 42, got %d", user.TeamID)
		}
	}
}

func TestTeamService_CreateTeam_Duplicate(t *testing.T) {
	teamRepo := &stubTeamRepo{teamExists: true}
	userRepo := &stubUserRepo{}
	svc := teamService{teamRepo: teamRepo, userRepo: userRepo}

	_, err := svc.CreateTeam("backend", nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrTeamExists) {
		t.Fatalf("expected ErrTeamExists, got %v", err)
	}
	if len(userRepo.created) != 0 {
		t.Fatalf("expected no users to be created when team exists")
	}
}

func TestTeamService_CreateTeam_UserRepoError(t *testing.T) {
	teamRepo := &stubTeamRepo{}
	userRepo := &stubUserRepo{createErr: errors.New("db error")}
	svc := teamService{teamRepo: teamRepo, userRepo: userRepo}

	_, err := svc.CreateTeam("backend", []model.User{{UserID: "u1"}})
	if err == nil {
		t.Fatalf("expected error from user repo, got nil")
	}
	if !errors.Is(err, userRepo.createErr) {
		t.Fatalf("expected user repo error, got %v", err)
	}
}

func TestTeamService_GetTeam_Success(t *testing.T) {
	expected := &model.Team{
		ID:   10,
		Name: "backend",
		Users: []model.User{
			{UserID: "u1"},
		},
	}
	teamRepo := &stubTeamRepo{getTeam: expected}
	userRepo := &stubUserRepo{}
	svc := teamService{teamRepo: teamRepo, userRepo: userRepo}

	team, err := svc.GetTeam("backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team != expected {
		t.Fatalf("expected pointer to team %p, got %p", expected, team)
	}
}

func TestTeamService_GetTeam_NotFound(t *testing.T) {
	teamRepo := &stubTeamRepo{getErr: repoerrs.ErrNotFound}
	userRepo := &stubUserRepo{}
	svc := teamService{teamRepo: teamRepo, userRepo: userRepo}

	_, err := svc.GetTeam("unknown")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestTeamService_GetTeam_RepoError(t *testing.T) {
	teamRepo := &stubTeamRepo{getErr: errors.New("db down")}
	userRepo := &stubUserRepo{}
	svc := teamService{teamRepo: teamRepo, userRepo: userRepo}

	_, err := svc.GetTeam("backend")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, teamRepo.getErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
