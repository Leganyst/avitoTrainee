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

// CreatePR godoc
// @Summary      Создать PR
// @Description  Создаёт PR и автоматически назначает доступных ревьюверов.
// @Tags         PullRequests
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreatePRRequest  true  "Данные PR"
// @Success      201      {object}  dto.CreatePRResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      409      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/pullRequest/create [post]
func (h *PRHandler) CreatePR(c *gin.Context) {
	log := logger(c)
	var req dto.CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnw("invalid create PR payload", "error", err)
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}

	if req.PRID == "" || req.Name == "" || req.Author == "" {
		log.Warnw("missing required PR fields", "payload", req)
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id, pull_request_name and author_id are required")
		return
	}
	log.Debugw("create PR request", "payload", req)

	pr, err := h.prSvc.CreatePR(req.PRID, req.Name, req.Author)
	if err != nil {
		log.Errorw("failed to create PR", "pr_id", req.PRID, "author", req.Author, "error", err)
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CreatePRResponse{
		PR: mapper.MapPullRequestToDTO(*pr),
	})
	log.Infow("PR created", "pr_id", pr.PRID, "author", req.Author, "reviewers", len(pr.AssignedReviewers))
}

// MergePR godoc
// @Summary      Merge PR
// @Description  Переводит PR в состояние MERGED (идемпотентно).
// @Tags         PullRequests
// @Accept       json
// @Produce      json
// @Param        request  body      dto.MergePRRequest  true  "Идентификатор PR"
// @Success      200      {object}  dto.MergePRResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/pullRequest/merge [post]
func (h *PRHandler) MergePR(c *gin.Context) {
	log := logger(c)
	var req dto.MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnw("invalid merge payload", "error", err)
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}
	if req.PRID == "" {
		log.Warnw("pull_request_id missing in merge request")
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id is required")
		return
	}
	log.Debugw("merge PR request", "payload", req)

	pr, err := h.prSvc.Merge(req.PRID)
	if err != nil {
		log.Errorw("failed to merge PR", "pr_id", req.PRID, "error", err)
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MergePRResponse{
		PR: mapper.MapPullRequestToDTO(*pr),
	})
	log.Infow("PR merged", "pr_id", pr.PRID)
}

// ReassignReviewer godoc
// @Summary      Переназначить ревьювера
// @Description  Заменяет ревьювера на другого активного участника его команды.
// @Tags         PullRequests
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ReassgnRequest  true  "Параметры переназначения"
// @Success      200      {object}  dto.ReassignResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      409      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/pullRequest/reassign [post]
func (h *PRHandler) ReassignReviewer(c *gin.Context) {
	log := logger(c)
	var req dto.ReassgnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnw("invalid reassign payload", "error", err)
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}
	if req.PRID == "" || req.OldUserID == "" {
		log.Warnw("missing fields for reassign", "payload", req)
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id and old_user_id are required")
		return
	}
	log.Debugw("reassign request", "payload", req)

	pr, replacedBy, err := h.prSvc.Reassign(req.PRID, req.OldUserID)
	if err != nil {
		log.Errorw("failed to reassign reviewer", "pr_id", req.PRID, "old_user", req.OldUserID, "error", err)
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ReassignResponse{
		PR:         mapper.MapPullRequestToDTO(*pr),
		ReplacedBy: replacedBy,
	})
	log.Infow("reviewer reassigned", "pr_id", pr.PRID, "old_user", req.OldUserID, "new_user", replacedBy)
}

func (h *PRHandler) handleError(c *gin.Context, err error) {
	log := logger(c)
	switch {
	case errors.Is(err, serviceerrs.ErrUserNotFound),
		errors.Is(err, serviceerrs.ErrPRNotFound):
		log.Warnw("resource not found", "error", err)
		writeError(c, http.StatusNotFound, errorCodeNotFound, err.Error())
	case errors.Is(err, serviceerrs.ErrPRExists):
		log.Warnw("PR already exists", "error", err)
		writeError(c, http.StatusConflict, errorCodePRExists, err.Error())
	case errors.Is(err, serviceerrs.ErrPRMerged):
		log.Warnw("operation on merged PR", "error", err)
		writeError(c, http.StatusConflict, errorCodePRMerged, err.Error())
	case errors.Is(err, serviceerrs.ErrReviewerMissing):
		log.Warnw("reviewer missing", "error", err)
		writeError(c, http.StatusConflict, errorCodeNotAssigned, err.Error())
	case errors.Is(err, serviceerrs.ErrNoCandidates):
		log.Warnw("no candidates for reassignment", "error", err)
		writeError(c, http.StatusConflict, errorCodeNoCandidate, err.Error())
	default:
		log.Errorw("internal PR handler error", "error", err)
		writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
	}
}
