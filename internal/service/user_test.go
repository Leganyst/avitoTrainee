package service

import (
	"errors"
	"testing"

	"github.com/Leganyst/avitoTrainee/internal/model"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
	serviceerrs "github.com/Leganyst/avitoTrainee/internal/service/errs"
)

type stubUserPRRepo struct {
	prs    []model.PullRequest
	prErr  error
	called bool
}

func (s *stubUserPRRepo) CreatePR(pr *model.PullRequest) error { return nil }
func (s *stubUserPRRepo) AddReviewers(pr *model.PullRequest, reviewers []model.User) error {
	return nil
}
func (s *stubUserPRRepo) ReplaceReviewer(pr *model.PullRequest, oldReviewerID, newReviewerID uint) error {
	return nil
}
func (s *stubUserPRRepo) GetPRByExternalID(prID string) (*model.PullRequest, error) { return nil, nil }
func (s *stubUserPRRepo) UpdatePR(pr *model.PullRequest) error                      { return nil }

func (s *stubUserPRRepo) GetPRsWhereReviewer(userID uint) ([]model.PullRequest, error) {
	s.called = true
	if s.prErr != nil {
		return nil, s.prErr
	}
	return s.prs, nil
}

func TestUserService_SetActive_Success(t *testing.T) {
	repo := &stubUserRepo{
		users: map[string]*model.User{
			"u1": {UserID: "u1", IsActive: false},
		},
	}
	svc := userService{userRepo: repo, prRepo: &stubUserPRRepo{}}

	user, err := svc.SetActive("u1", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !user.IsActive {
		t.Fatalf("expected user to be active")
	}
}

func TestUserService_SetActive_NotFound(t *testing.T) {
	repo := &stubUserRepo{
		setActiveErr: repoerrs.ErrNotFound,
	}
	svc := userService{userRepo: repo, prRepo: &stubUserPRRepo{}}

	_, err := svc.SetActive("missing", true)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	expected := &model.User{UserID: "u1"}
	repo := &stubUserRepo{
		users: map[string]*model.User{"u1": expected},
	}
	svc := userService{userRepo: repo, prRepo: &stubUserPRRepo{}}

	user, err := svc.GetUserByID("u1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user != expected {
		t.Fatalf("expected pointer %p, got %p", expected, user)
	}
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	repo := &stubUserRepo{
		getErrFor: map[string]error{"missing": repoerrs.ErrNotFound},
	}
	svc := userService{userRepo: repo, prRepo: &stubUserPRRepo{}}

	_, err := svc.GetUserByID("missing")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_GetUserReviews_Success(t *testing.T) {
	user := &model.User{ID: 10, UserID: "u1"}
	userRepo := &stubUserRepo{users: map[string]*model.User{"u1": user}}
	prRepo := &stubUserPRRepo{
		prs: []model.PullRequest{{PRID: "pr-1"}},
	}
	svc := userService{userRepo: userRepo, prRepo: prRepo}

	prs, err := svc.GetUserReviews("u1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(prs) != 1 || prs[0].PRID != "pr-1" {
		t.Fatalf("unexpected PRs result: %+v", prs)
	}
	if !prRepo.called {
		t.Fatalf("expected PR repository to be called")
	}
}

func TestUserService_GetUserReviews_UserNotFound(t *testing.T) {
	userRepo := &stubUserRepo{
		getErrFor: map[string]error{"missing": repoerrs.ErrNotFound},
	}
	prRepo := &stubUserPRRepo{}
	svc := userService{userRepo: userRepo, prRepo: prRepo}

	_, err := svc.GetUserReviews("missing")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, serviceerrs.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if prRepo.called {
		t.Fatalf("expected PR repo not to be called when user missing")
	}
}

func TestUserService_GetUserReviews_PRRepoError(t *testing.T) {
	user := &model.User{ID: 10, UserID: "u1"}
	userRepo := &stubUserRepo{users: map[string]*model.User{"u1": user}}
	prRepo := &stubUserPRRepo{prErr: errors.New("db error")}
	svc := userService{userRepo: userRepo, prRepo: prRepo}

	_, err := svc.GetUserReviews("u1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, prRepo.prErr) {
		t.Fatalf("expected PR repo error, got %v", err)
	}
}
