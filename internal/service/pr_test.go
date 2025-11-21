package service

import (
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/Leganyst/avitoTrainee/internal/model"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
	serviceerrs "github.com/Leganyst/avitoTrainee/internal/service/errs"
)

func TestPRService_Merge_SetsStatusMergedAndTimestamp(t *testing.T) {
	pr := &model.PullRequest{PRID: "pr-1", Status: statusOpen}
	repo := &stubPRRepo{pr: pr}
	svc := prService{repo: repo, userRepo: &stubUserRepo{}}

	got, err := svc.Merge("pr-1")
	if err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}
	if got.Status != statusMerged {
		t.Fatalf("expected status %q, got %q", statusMerged, got.Status)
	}
	if got.UpdatedAt == nil || got.UpdatedAt.After(time.Now()) {
		t.Fatalf("expected UpdatedAt to be set, got %v", got.UpdatedAt)
	}
	if !repo.updateCalled {
		t.Fatalf("expected UpdatePR to be called")
	}
}

func TestPRService_Merge_Idempotent(t *testing.T) {
	now := time.Now()
	pr := &model.PullRequest{PRID: "pr-merged", Status: statusMerged, UpdatedAt: &now}
	repo := &stubPRRepo{pr: pr}
	svc := prService{repo: repo, userRepo: &stubUserRepo{}}

	got, err := svc.Merge("pr-merged")
	if err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}
	if got.Status != statusMerged {
		t.Fatalf("expected status to stay %q, got %q", statusMerged, got.Status)
	}
	if repo.updateCalled {
		t.Fatalf("expected no repository update on idempotent merge")
	}
}

func TestPRService_Merge_NotFound(t *testing.T) {
	repo := &stubPRRepo{getErr: repoerrs.ErrNotFound}
	svc := prService{repo: repo, userRepo: &stubUserRepo{}}

	_, err := svc.Merge("missing")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrPRNotFound) {
		t.Fatalf("expected ErrPRNotFound, got %v", err)
	}
}

func TestPRService_CreatePR_Success(t *testing.T) {
	prevRnd := rnd
	rnd = rand.New(rand.NewSource(1))
	defer func() { rnd = prevRnd }()

	userRepo := &stubUserRepo{
		users: map[string]*model.User{
			"author": {ID: 1, UserID: "author", TeamID: 10},
		},
		activeByTeam: map[uint][]model.User{
			10: {
				{ID: 2, UserID: "u2", TeamID: 10},
				{ID: 3, UserID: "u3", TeamID: 10},
			},
		},
	}
	prRepo := &stubPRRepo{}
	svc := prService{repo: prRepo, userRepo: userRepo}

	pr, err := svc.CreatePR("pr-1", "New feature", "author")
	if err != nil {
		t.Fatalf("CreatePR returned error: %v", err)
	}
	if pr.Status != statusOpen || pr.AuthorID != "author" {
		t.Fatalf("unexpected PR fields: %+v", pr)
	}
	if len(pr.AssignedReviewers) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(pr.AssignedReviewers))
	}
	if !prRepo.addReviewersCall {
		t.Fatalf("expected AddReviewers call")
	}
}

func TestPRService_CreatePR_UserNotFound(t *testing.T) {
	userRepo := &stubUserRepo{
		getErrFor: map[string]error{"author": repoerrs.ErrNotFound},
	}
	prRepo := &stubPRRepo{}
	svc := prService{repo: prRepo, userRepo: userRepo}

	_, err := svc.CreatePR("pr-1", "New feature", "author")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestPRService_CreatePR_Duplicate(t *testing.T) {
	userRepo := &stubUserRepo{
		users: map[string]*model.User{
			"author": {ID: 1, UserID: "author", TeamID: 10},
		},
	}
	prRepo := &stubPRRepo{createErr: repoerrs.ErrDuplicate}
	svc := prService{repo: prRepo, userRepo: userRepo}

	_, err := svc.CreatePR("pr-1", "New feature", "author")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrPRExists) {
		t.Fatalf("expected ErrPRExists, got %v", err)
	}
}

func TestPRService_Reassign_Success(t *testing.T) {
	pr := &model.PullRequest{
		PRID:   "pr-1",
		Status: statusOpen,
		AssignedReviewers: []model.User{
			{ID: 2, UserID: "u2", TeamID: 20},
			{ID: 3, UserID: "u3", TeamID: 20},
		},
	}
	userRepo := &stubUserRepo{
		users: map[string]*model.User{
			"u2": {ID: 2, UserID: "u2", TeamID: 20},
		},
		activeByTeam: map[uint][]model.User{
			20: {{ID: 4, UserID: "u4", TeamID: 20}},
		},
	}
	prRepo := &stubPRRepo{pr: pr}
	svc := prService{repo: prRepo, userRepo: userRepo}

	result, replacedBy, err := svc.Reassign("pr-1", "u2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if replacedBy != "u4" {
		t.Fatalf("expected replacement u4, got %s", replacedBy)
	}
	found := false
	for _, reviewer := range result.AssignedReviewers {
		if reviewer.ID == 4 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected assigned reviewers to include replacement")
	}
	if !prRepo.replacedCalled || prRepo.replaceOldID != 2 || prRepo.replaceNewID != 4 {
		t.Fatalf("expected ReplaceReviewer to be called with correct IDs")
	}
}

func TestPRService_Reassign_Merged(t *testing.T) {
	pr := &model.PullRequest{PRID: "pr-1", Status: statusMerged}
	prRepo := &stubPRRepo{pr: pr}
	svc := prService{repo: prRepo, userRepo: &stubUserRepo{}}

	_, _, err := svc.Reassign("pr-1", "u2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrPRMerged) {
		t.Fatalf("expected ErrPRMerged, got %v", err)
	}
}

func TestPRService_Reassign_ReviewerMissing(t *testing.T) {
	pr := &model.PullRequest{
		PRID:   "pr-1",
		Status: statusOpen,
		AssignedReviewers: []model.User{
			{ID: 3, UserID: "u3", TeamID: 20},
		},
	}
	userRepo := &stubUserRepo{
		users: map[string]*model.User{
			"u2": {ID: 2, UserID: "u2", TeamID: 20},
		},
	}
	prRepo := &stubPRRepo{pr: pr}
	svc := prService{repo: prRepo, userRepo: userRepo}

	_, _, err := svc.Reassign("pr-1", "u2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrReviewerMissing) {
		t.Fatalf("expected ErrReviewerMissing, got %v", err)
	}
}

func TestPRService_Reassign_NoCandidates(t *testing.T) {
	pr := &model.PullRequest{
		PRID:   "pr-1",
		Status: statusOpen,
		AssignedReviewers: []model.User{
			{ID: 2, UserID: "u2", TeamID: 20},
		},
	}
	userRepo := &stubUserRepo{
		users: map[string]*model.User{
			"u2": {ID: 2, UserID: "u2", TeamID: 20},
		},
		activeByTeam: map[uint][]model.User{
			20: {
				{ID: 2, UserID: "u2", TeamID: 20}, // excluded
			},
		},
	}
	prRepo := &stubPRRepo{pr: pr}
	svc := prService{repo: prRepo, userRepo: userRepo}

	_, _, err := svc.Reassign("pr-1", "u2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrNoCandidates) {
		t.Fatalf("expected ErrNoCandidates, got %v", err)
	}
}
