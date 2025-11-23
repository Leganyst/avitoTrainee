package service

import (
	"github.com/Leganyst/avitoTrainee/internal/model"
	repoerrs "github.com/Leganyst/avitoTrainee/internal/repository/errs"
)

// ----- User repository stub -----
type stubUserRepo struct {
	users        map[string]*model.User
	getErrFor    map[string]error
	createErr    error
	setActiveErr error
	activeByTeam map[uint][]model.User
	usersByTeam  map[uint][]model.User
	activeErr    error
	created      []model.User
	bulkErr      error
}

func (s *stubUserRepo) CreateOrUpdate(user *model.User) error {
	if s.createErr != nil {
		return s.createErr
	}
	cpy := *user
	s.created = append(s.created, cpy)
	return nil
}
func (s *stubUserRepo) GetByUserID(userID string) (*model.User, error) {
	if s.getErrFor != nil {
		if err, ok := s.getErrFor[userID]; ok {
			return nil, err
		}
	}
	if s.users != nil {
		if u, ok := s.users[userID]; ok {
			return u, nil
		}
	}
	return nil, repoerrs.ErrNotFound
}
func (s *stubUserRepo) GetUsersByTeam(teamID uint) ([]model.User, error) {
	if s.usersByTeam == nil {
		return nil, nil
	}
	users := s.usersByTeam[teamID]
	cpy := make([]model.User, len(users))
	copy(cpy, users)
	return cpy, nil
}
func (s *stubUserRepo) SetActive(userID string, active bool) (*model.User, error) {
	if s.setActiveErr != nil {
		return nil, s.setActiveErr
	}
	u, ok := s.users[userID]
	if !ok {
		return nil, repoerrs.ErrNotFound
	}
	cpy := *u
	cpy.IsActive = active
	return &cpy, nil
}
func (s *stubUserRepo) GetActiveUsersByTeam(teamID uint) ([]model.User, error) {
	if s.activeErr != nil {
		return nil, s.activeErr
	}
	users := s.activeByTeam[teamID]
	cpy := make([]model.User, len(users))
	copy(cpy, users)
	return cpy, nil
}
func (s *stubUserRepo) BulkDeactivate(teamID uint, userIDs []string) ([]model.User, error) {
	if s.bulkErr != nil {
		return nil, s.bulkErr
	}
	if len(userIDs) == 0 {
		return nil, nil
	}

	idSet := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		idSet[id] = struct{}{}
	}

	var res []model.User
	for _, u := range s.users {
		if u.TeamID != teamID {
			continue
		}
		if _, ok := idSet[u.UserID]; !ok {
			continue
		}
		if !u.IsActive {
			continue
		}
		cpy := *u
		cpy.IsActive = false
		res = append(res, cpy)
	}

	if len(res) == 0 {
		return nil, repoerrs.ErrNotFound
	}
	return res, nil
}

// ----- PR repository stub -----
type stubPRRepo struct {
	pr               *model.PullRequest
	getErr           error
	updateErr        error
	updateCalled     bool
	createErr        error
	createdPR        *model.PullRequest
	addReviewersErr  error
	addedReviewers   []model.User
	replaceErr       error
	replaceOldID     uint
	replaceNewID     uint
	replacedCalled   bool
	addReviewersCall bool
	prsByReviewer    []model.PullRequest
	prsErr           error
	openPRs          []model.PullRequest
	openPRsErr       error
}

func (s *stubPRRepo) CreatePR(pr *model.PullRequest) error {
	if s.createErr != nil {
		return s.createErr
	}
	s.createdPR = pr
	if s.pr == nil {
		s.pr = pr
	}
	return nil
}
func (s *stubPRRepo) AddReviewers(pr *model.PullRequest, reviewers []model.User) error {
	if s.addReviewersErr != nil {
		return s.addReviewersErr
	}
	s.addReviewersCall = true
	s.addedReviewers = append([]model.User(nil), reviewers...)
	return nil
}
func (s *stubPRRepo) ReplaceReviewer(pr *model.PullRequest, oldReviewerID uint, newReviewer model.User) error {
	if s.replaceErr != nil {
		return s.replaceErr
	}
	s.replacedCalled = true
	s.replaceOldID = oldReviewerID
	s.replaceNewID = newReviewer.ID
	return nil
}
func (s *stubPRRepo) GetPRByExternalID(prID string) (*model.PullRequest, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.pr, nil
}
func (s *stubPRRepo) UpdatePR(pr *model.PullRequest) error {
	s.updateCalled = true
	if s.updateErr != nil {
		return s.updateErr
	}
	s.pr.Status = pr.Status
	s.pr.UpdatedAt = pr.UpdatedAt
	return nil
}
func (s *stubPRRepo) GetPRsWhereReviewer(userID uint) ([]model.PullRequest, error) {
	if s.prsErr != nil {
		return nil, s.prsErr
	}
	cpy := make([]model.PullRequest, len(s.prsByReviewer))
	copy(cpy, s.prsByReviewer)
	return cpy, nil
}
func (s *stubPRRepo) GetOpenPRsByReviewerIDs(reviewerIDs []uint) ([]model.PullRequest, error) {
	if s.openPRsErr != nil {
		return nil, s.openPRsErr
	}
	cpy := make([]model.PullRequest, len(s.openPRs))
	copy(cpy, s.openPRs)
	return cpy, nil
}

// ----- Team repository stub -----
type stubTeamRepo struct {
	teamExists bool
	createErr  error
	getTeam    *model.Team
	getErr     error
}

func (s *stubTeamRepo) CreateTeam(team *model.Team) error {
	if s.createErr != nil {
		return s.createErr
	}
	if team.ID == 0 {
		team.ID = 42
	}
	return nil
}
func (s *stubTeamRepo) GetTeamByName(name string) (*model.Team, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.getTeam, nil
}
func (s *stubTeamRepo) TeamExists(name string) (bool, error) { return s.teamExists, nil }
