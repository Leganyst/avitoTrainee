package test

import (
	"errors"
	"testing"

	"github.com/Leganyst/avitoTrainee/internal/model"
	"github.com/Leganyst/avitoTrainee/internal/repository"
	"github.com/Leganyst/avitoTrainee/internal/service"
	serviceerrs "github.com/Leganyst/avitoTrainee/internal/service/errs"
)

func TestPRFlow_Create_Reassign_Merge(t *testing.T) {
	db := connectTestDB(t)
	prepareDB(t, db)

	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)

	teamSvc := service.NewTeamService(repository.NewTeamRepository(db), userRepo)
	prSvc := service.NewPrService(prRepo, userRepo)

	members := []model.User{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
		{UserID: "u3", Username: "Charlie", IsActive: true},
		{UserID: "u4", Username: "Oleg", IsActive: true},
	}

	if _, err := teamSvc.CreateTeam("backend", members); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	// act: create PR
	pr, err := prSvc.CreatePR("pr-1", "Add search", "u1")
	if err != nil {
		t.Fatalf("CreatePR returned error: %v", err)
	}

	// assert: status и ревьюверы
	if pr.Status != "OPEN" {
		t.Fatalf("expected status OPEN, got %s", pr.Status)
	}
	if len(pr.AssignedReviewers) == 0 || len(pr.AssignedReviewers) > 2 {
		t.Fatalf("expected 1..2 reviewers, got %d", len(pr.AssignedReviewers))
	}
	for _, r := range pr.AssignedReviewers {
		if r.UserID == "u1" {
			t.Fatalf("author must not be assigned as reviewer")
		}
	}

	// act: reassign одного ревьювера (кандидат - u4)
	old := pr.AssignedReviewers[0].UserID
	updated, replacedBy, err := prSvc.Reassign(pr.PRID, old)
	if err != nil {
		t.Fatalf("Reassign returned error: %v", err)
	}
	if replacedBy == old {
		t.Fatalf("replace must differ from old reviewer")
	}
	found := false
	for _, r := range updated.AssignedReviewers {
		if r.UserID == replacedBy {
			found = true
		}
		if r.UserID == old {
			t.Fatalf("old reviewer should be replaced")
		}
	}
	if !found {
		t.Fatalf("new reviewer not present in AssignedReviewers")
	}

	// act: merge (идемпотентность проверим повторным вызовом)
	merged, err := prSvc.Merge(pr.PRID)
	if err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}
	if merged.Status != "MERGED" {
		t.Fatalf("expected MERGED, got %s", merged.Status)
	}
	if _, err := prSvc.Merge(pr.PRID); err != nil {
		t.Fatalf("second Merge should be idempotent, got %v", err)
	}

	// act: попытка reassign после MERGED
	if _, _, err := prSvc.Reassign(pr.PRID, merged.AssignedReviewers[0].UserID); !errors.Is(err, serviceerrs.ErrPRMerged) {
		t.Fatalf("expected ErrPRMerged after merge, got %v", err)
	}
}

func TestPRFlow_Create_NoAuthor(t *testing.T) {
	db := connectTestDB(t)
	prepareDB(t, db)

	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)

	prSvc := service.NewPrService(prRepo, userRepo)

	if _, err := prSvc.CreatePR("pr-x", "Feature", "missing"); !errors.Is(err, serviceerrs.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestPRFlow_Create_Duplicate(t *testing.T) {
	db := connectTestDB(t)
	prepareDB(t, db)

	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)

	teamSvc := service.NewTeamService(repository.NewTeamRepository(db), userRepo)
	prSvc := service.NewPrService(prRepo, userRepo)

	_, _ = teamSvc.CreateTeam("backend", []model.User{{UserID: "u1", Username: "Alice", IsActive: true}})
	if _, err := prSvc.CreatePR("pr-1", "Feature", "u1"); err != nil {
		t.Fatalf("first CreatePR err: %v", err)
	}
	// повтор создания PR
	if _, err := prSvc.CreatePR("pr-1", "Feature", "u1"); !errors.Is(err, serviceerrs.ErrPRExists) {
		t.Fatalf("expected ErrPRExists, got %v", err)
	}
}

func TestPRFlow_Reassign_NoCandidates(t *testing.T) {
	db := connectTestDB(t)
	prepareDB(t, db)

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)

	teamSvc := service.NewTeamService(teamRepo, userRepo)
	prSvc := service.NewPrService(prRepo, userRepo)

	// только один активный кроме автора -> кандидатов нет
	_, _ = teamSvc.CreateTeam("backend", []model.User{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
	})
	pr, err := prSvc.CreatePR("pr-1", "Feature", "u1")
	if err != nil {
		t.Fatalf("CreatePR err: %v", err)
	}
	if _, _, err := prSvc.Reassign(pr.PRID, pr.AssignedReviewers[0].UserID); !errors.Is(err, serviceerrs.ErrNoCandidates) {
		t.Fatalf("expected ErrNoCandidates, got %v", err)
	}
}

func TestPRFlow_Reassign_ReviewerMissing(t *testing.T) {
	db := connectTestDB(t)
	prepareDB(t, db)

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)

	teamSvc := service.NewTeamService(teamRepo, userRepo)
	prSvc := service.NewPrService(prRepo, userRepo)

	_, _ = teamSvc.CreateTeam("backend", []model.User{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
		{UserID: "u3", Username: "Eve", IsActive: true},
	})
	pr, err := prSvc.CreatePR("pr-1", "Feature", "u1")
	if err != nil {
		t.Fatalf("CreatePR err: %v", err)
	}
	if _, _, err := prSvc.Reassign(pr.PRID, "unknown"); !errors.Is(err, serviceerrs.ErrReviewerMissing) {
		t.Fatalf("expected ErrReviewerMissing, got %v", err)
	}
}

func TestPR_Reassign_To_Merged_PR(t *testing.T) {
	db := connectTestDB(t)
	prepareDB(t, db)

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)

	teamSvc := service.NewTeamService(teamRepo, userRepo)
	prSvc := service.NewPrService(prRepo, userRepo)

	_, _ = teamSvc.CreateTeam("backend", []model.User{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
		{UserID: "u3", Username: "Eve", IsActive: true},
	})

	pr, err := prSvc.CreatePR("pr-1", "Feature", "u1")
	if err != nil {
		t.Fatalf("CreatePR err: %v", err)
	}

	pr, err = prSvc.Merge(pr.PRID)
	if err != nil {
		t.Fatalf("Merge PR err: %v", err)
	}

	if _, _, err = prSvc.Reassign(pr.PRID, pr.AssignedReviewers[0].UserID); !errors.Is(err, serviceerrs.ErrPRMerged) {
		t.Fatalf("This PR not reassigned viewers, err: %v", err)
	}
}
