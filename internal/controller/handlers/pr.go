package handlers

import (
	"errors"
	"net/http"

	"github.com/Leganyst/avitoTrainee/internal/controller/dto"
	"github.com/Leganyst/avitoTrainee/internal/mapper"
	"github.com/Leganyst/avitoTrainee/internal/service"
	serviceerrs "github.com/Leganyst/avitoTrainee/internal/service/errs"
	"github.com/gin-gonic/gin"
)

type PRHandler struct {
	prSvc service.PRService
}

func NewPRHandler(prSvc service.PRService) *PRHandler {
	return &PRHandler{prSvc: prSvc}
}

func registerPRRoutes(r gin.IRouter, prSvc service.PRService) {
	handler := NewPRHandler(prSvc)

	group := r.Group("/pullRequest")
	group.POST("/create", handler.CreatePR)
	group.POST("/merge", handler.MergePR)
	group.POST("/reassign", handler.ReassignReviewer)
}

// CreatePR создаёт PR и назначает ревьюверов согласно бизнес-правилам.
func (h *PRHandler) CreatePR(c *gin.Context) {
	var req dto.CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}

	if req.PRID == "" || req.Name == "" || req.Author == "" {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id, pull_request_name and author_id are required")
		return
	}

	pr, err := h.prSvc.CreatePR(req.PRID, req.Name, req.Author)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CreatePRResponse{
		PR: mapper.MapPullRequestToDTO(*pr),
	})
}

// MergePR помечает PR как MERGED (идемпотентно).
func (h *PRHandler) MergePR(c *gin.Context) {
	var req dto.MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}
	if req.PRID == "" {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id is required")
		return
	}

	pr, err := h.prSvc.Merge(req.PRID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MergePRResponse{
		PR: mapper.MapPullRequestToDTO(*pr),
	})
}

// ReassignReviewer заменяет ревьювера на другого из его команды.
func (h *PRHandler) ReassignReviewer(c *gin.Context) {
	var req dto.ReassgnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}
	if req.PRID == "" || req.OldUserID == "" {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id and old_user_id are required")
		return
	}

	pr, replacedBy, err := h.prSvc.Reassign(req.PRID, req.OldUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ReassignResponse{
		PR:         mapper.MapPullRequestToDTO(*pr),
		ReplacedBy: replacedBy,
	})
}

func (h *PRHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, serviceerrs.ErrUserNotFound),
		errors.Is(err, serviceerrs.ErrPRNotFound):
		writeError(c, http.StatusNotFound, errorCodeNotFound, err.Error())
	case errors.Is(err, serviceerrs.ErrPRExists):
		writeError(c, http.StatusConflict, errorCodePRExists, err.Error())
	case errors.Is(err, serviceerrs.ErrPRMerged):
		writeError(c, http.StatusConflict, errorCodePRMerged, err.Error())
	case errors.Is(err, serviceerrs.ErrReviewerMissing):
		writeError(c, http.StatusConflict, errorCodeNotAssigned, err.Error())
	case errors.Is(err, serviceerrs.ErrNoCandidates):
		writeError(c, http.StatusConflict, errorCodeNoCandidate, err.Error())
	default:
		writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
	}
}
