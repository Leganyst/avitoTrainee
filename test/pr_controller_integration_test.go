package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
)

func TestPRController_FullFlow(t *testing.T) {
	server := newAPITestServer(t)

	createTeamPayload := `{
		"team_name": "backend",
		"members": [
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "Oleg", "is_active": true}
		]
	}`
	resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/team/add", createTeamPayload))
	if resp.Code != http.StatusCreated {
		t.Fatalf("create team status = %d, want %d", resp.Code, http.StatusCreated)
	}

	createPRPayload := `{"pull_request_id": "pr-1", "pull_request_name": "Add search", "author_id": "u1"}`
	resp = server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/create", createPRPayload))
	if resp.Code != http.StatusCreated {
		t.Fatalf("create PR status = %d, want %d", resp.Code, http.StatusCreated)
	}

	createBody := decodeBody[dto.CreatePRResponse](t, resp.Body)
	if createBody.PR.Status != "OPEN" {
		t.Fatalf("PR status = %s, want OPEN", createBody.PR.Status)
	}
	if len(createBody.PR.AssignedReviewers) == 0 || len(createBody.PR.AssignedReviewers) > 2 {
		t.Fatalf("expected 1..2 reviewers, got %d", len(createBody.PR.AssignedReviewers))
	}
	for _, reviewer := range createBody.PR.AssignedReviewers {
		if reviewer == "u1" {
			t.Fatalf("author must not be assigned as reviewer")
		}
	}

	oldReviewer := createBody.PR.AssignedReviewers[0]
	reassignPayload := fmt.Sprintf(`{"pull_request_id": "%s", "old_user_id": "%s"}`, createBody.PR.PRID, oldReviewer)
	resp = server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/reassign", reassignPayload))
	if resp.Code != http.StatusOK {
		t.Fatalf("reassign status = %d, want %d", resp.Code, http.StatusOK)
	}

	reassignBody := decodeBody[dto.ReassignResponse](t, resp.Body)
	if reassignBody.ReplacedBy == oldReviewer {
		t.Fatalf("reviewer must change, got the same id %s", reassignBody.ReplacedBy)
	}
	found := false
	for _, reviewer := range reassignBody.PR.AssignedReviewers {
		if reviewer == reassignBody.ReplacedBy {
			found = true
		}
		if reviewer == oldReviewer {
			t.Fatalf("old reviewer should be removed")
		}
	}
	if !found {
		t.Fatalf("new reviewer %s not in list", reassignBody.ReplacedBy)
	}

	mergePayload := fmt.Sprintf(`{"pull_request_id": "%s"}`, createBody.PR.PRID)
	resp = server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/merge", mergePayload))
	if resp.Code != http.StatusOK {
		t.Fatalf("merge status = %d, want %d", resp.Code, http.StatusOK)
	}

	mergeBody := decodeBody[dto.MergePRResponse](t, resp.Body)
	if mergeBody.PR.Status != "MERGED" {
		t.Fatalf("PR status after merge = %s, want MERGED", mergeBody.PR.Status)
	}

	// повторный reassign после MERGED должен вернуть конфликт с кодом PR_MERGED
	conflictPayload := fmt.Sprintf(`{"pull_request_id": "%s", "old_user_id": "%s"}`, createBody.PR.PRID, mergeBody.PR.AssignedReviewers[0])
	resp = server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/reassign", conflictPayload))
	if resp.Code != http.StatusConflict {
		t.Fatalf("reassign on merged status = %d, want %d", resp.Code, http.StatusConflict)
	}

	errBody := decodeBody[dto.ErrorResponse](t, resp.Body)
	if errBody.Error.Code != "PR_MERGED" {
		t.Fatalf("error code = %s, want PR_MERGED", errBody.Error.Code)
	}
}

func newJSONRequest(t *testing.T, method, url, body string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeBody[T any](t *testing.T, body io.Reader) T {
	t.Helper()

	var target T
	if err := json.NewDecoder(body).Decode(&target); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return target
}

func assertErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, status int, code string) dto.ErrorResponse {
	t.Helper()

	if resp.Code != status {
		t.Fatalf("status = %d, want %d", resp.Code, status)
	}
	errBody := decodeBody[dto.ErrorResponse](t, resp.Body)
	if errBody.Error.Code != code {
		t.Fatalf("error code = %s, want %s", errBody.Error.Code, code)
	}
	return errBody
}

func TestPRController_Create_Errors(t *testing.T) {
	server := newAPITestServer(t)

	t.Run("invalid json", func(t *testing.T) {
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/create", "{}"))
		body := assertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
		if body.Error.Message != "invalid request payload" {
			t.Fatalf("unexpected error message: %s", body.Error.Message)
		}
	})

	t.Run("author not found", func(t *testing.T) {
		payload := `{"pull_request_id": "pr-404", "pull_request_name": "Feature", "author_id": "unknown"}`
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/create", payload))
		assertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
	})

	t.Run("duplicate PR", func(t *testing.T) {
		teamPayload := `{"team_name":"backend","members":[{"user_id":"u1","username":"Alice","is_active":true}]}`
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/team/add", teamPayload))
		if resp.Code != http.StatusCreated {
			t.Fatalf("failed to create team, status %d", resp.Code)
		}

		createPayload := `{"pull_request_id":"pr-dup","pull_request_name":"Feature","author_id":"u1"}`
		resp = server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/create", createPayload))
		if resp.Code != http.StatusCreated {
			t.Fatalf("first create status %d", resp.Code)
		}
		resp = server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/create", createPayload))
		assertErrorResponse(t, resp, http.StatusConflict, "PR_EXISTS")
	})
}

func TestPRController_Merge_Errors(t *testing.T) {
	server := newAPITestServer(t)

	t.Run("invalid json", func(t *testing.T) {
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/merge", "{}"))
		body := assertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
		if body.Error.Message != "invalid request payload" {
			t.Fatalf("unexpected error message: %s", body.Error.Message)
		}
	})

	t.Run("pr not found", func(t *testing.T) {
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/merge", `{"pull_request_id":"missing"}`))
		assertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
	})
}

func TestPRController_Reassign_Errors(t *testing.T) {
	server := newAPITestServer(t)

	t.Run("invalid json", func(t *testing.T) {
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/reassign", "{}"))
		assertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
	})

	t.Run("pr not found", func(t *testing.T) {
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/reassign", `{"pull_request_id":"missing","old_user_id":"u1"}`))
		assertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
	})

	t.Run("reviewer missing", func(t *testing.T) {
		setupTeamAndPR(t, server, "team-rm", "u-auth", 3)
		payload := `{"pull_request_id":"pr-1","old_user_id":"unknown"}`
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/reassign", payload))
		assertErrorResponse(t, resp, http.StatusConflict, "NOT_ASSIGNED")
	})

	t.Run("no candidates for reassignment", func(t *testing.T) {
		server := newAPITestServer(t)
		teamPayload := `{"team_name":"small","members":[{"user_id":"u1","username":"Alice","is_active":true},{"user_id":"u2","username":"Bob","is_active":true}]}`
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/team/add", teamPayload))
		if resp.Code != http.StatusCreated {
			t.Fatalf("create team status %d", resp.Code)
		}
		createPayload := `{"pull_request_id":"pr-1","pull_request_name":"Feature","author_id":"u1"}`
		resp = server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/create", createPayload))
		if resp.Code != http.StatusCreated {
			t.Fatalf("create PR status %d", resp.Code)
		}
		payload := `{"pull_request_id":"pr-1","old_user_id":"u2"}`
		resp = server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/reassign", payload))
		assertErrorResponse(t, resp, http.StatusConflict, "NO_CANDIDATE")
	})
}

// setupTeamAndPR создаёт команду с указанным размером и один PR с автором.
func setupTeamAndPR(t *testing.T, server *apiTestServer, teamName, authorID string, members int) {
	t.Helper()

	teamReq := dto.CreateTeamRequest{
		TeamName: teamName,
		Members:  make([]dto.TeamMember, 0, members),
	}
	for i := 0; i < members; i++ {
		id := fmt.Sprintf("%s-%d", authorID, i)
		if i == 0 {
			id = authorID
		}
		teamReq.Members = append(teamReq.Members, dto.TeamMember{
			UserID:   id,
			Username: "user-" + strconv.Itoa(i),
			IsActive: true,
		})
	}
	body, _ := json.Marshal(teamReq)
	resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/team/add", string(body)))
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to create team, status %d", resp.Code)
	}

	prReq := dto.CreatePRRequest{
		PRID:   "pr-1",
		Name:   "Feature",
		Author: authorID,
	}
	body, _ = json.Marshal(prReq)
	resp = server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/create", string(body)))
	if resp.Code != http.StatusCreated {
		t.Fatalf("failed to create PR, status %d", resp.Code)
	}
}

// Нагрузочный тест
func TestPRController_SLIsUnderModerateLoad(t *testing.T) {
	server := newAPITestServer(t)

	const (
		teamsCount   = 20
		membersCount = 10
		maxLatency   = 300 * time.Millisecond
		targetRPS    = 5
	)

	// Готовим 20 команд по 10 человек (всего 200 пользователей).
	for teamIdx := 1; teamIdx <= teamsCount; teamIdx++ {
		payload := dto.CreateTeamRequest{
			TeamName: fmt.Sprintf("team-%02d", teamIdx),
			Members:  make([]dto.TeamMember, 0, membersCount),
		}
		for userIdx := 1; userIdx <= membersCount; userIdx++ {
			payload.Members = append(payload.Members, dto.TeamMember{
				UserID:   fmt.Sprintf("u-%02d-%02d", teamIdx, userIdx),
				Username: fmt.Sprintf("User-%02d-%02d", teamIdx, userIdx),
				IsActive: true,
			})
		}
		body, _ := json.Marshal(payload)
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/team/add", string(body)))
		if resp.Code != http.StatusCreated {
			t.Fatalf("create team %d status = %d, want %d", teamIdx, resp.Code, http.StatusCreated)
		}
	}

	// Подготовим 25 запросов создания PR и будем слать их с темпом 5 RPS.
	requests := make([]string, 0, 25)
	for i := 1; i <= 25; i++ {
		teamIdx := (i % teamsCount) + 1 // равномерно по командам
		payload := dto.CreatePRRequest{
			PRID:   fmt.Sprintf("pr-%03d", i),
			Name:   fmt.Sprintf("Feature-%03d", i),
			Author: fmt.Sprintf("u-%02d-01", teamIdx),
		}
		body, _ := json.Marshal(payload)
		requests = append(requests, string(body))
	}

	ticker := time.NewTicker(time.Second / time.Duration(targetRPS))
	defer ticker.Stop()

	var (
		successCount int
		maxDuration  time.Duration
	)

	for idx, body := range requests {
		<-ticker.C

		start := time.Now()
		resp := server.doRequest(newJSONRequest(t, http.MethodPost, "/api/pullRequest/create", body))
		duration := time.Since(start)

		if duration > maxDuration {
			maxDuration = duration
		}
		if resp.Code >= 200 && resp.Code < 300 {
			successCount++
		} else {
			t.Fatalf("request %d failed: status %d", idx, resp.Code)
		}
	}

	successRate := float64(successCount) / float64(len(requests))
	if successRate < 0.999 {
		t.Fatalf("success rate %.4f < 0.999", successRate)
	}
	if maxDuration > maxLatency {
		t.Fatalf("max latency %s > %s", maxDuration, maxLatency)
	}
}
