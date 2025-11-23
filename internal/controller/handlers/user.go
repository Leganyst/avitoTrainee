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

type UserHandler struct {
	userSvc service.UserService
}

func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func registerUserRoutes(r gin.IRouter, userSvc service.UserService) {
	handler := NewUserHandler(userSvc)

	group := r.Group("/users")
	group.POST("/setIsActive", handler.SetActive)
	group.GET("/getReview", handler.GetUserReviews)
	group.POST("/bulkDeactivate", handler.BulkDeactivate)
}

// SetActive godoc
// @Summary      Обновить активность пользователя
// @Description  Ставит или снимает флаг активности пользователя.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UserRequest  true  "Параметры активности"
// @Success      200      {object}  dto.UserResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/users/setIsActive [post]
func (h *UserHandler) SetActive(c *gin.Context) {
	log := logger(c)
	var req dto.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnw("invalid SetActive payload", "error", err)
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}
	log.Debugw("set active request", "payload", req)

	user, err := h.userSvc.SetActive(req.UserID, *req.IsActive)
	if err != nil {
		log.Errorw("failed to update user activity", "user_id", req.UserID, "is_active", req.IsActive, "error", err)
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		User: mapper.MapUserToDTO(*user),
	})
	log.Infow("user activity updated", "user_id", req.UserID, "is_active", req.IsActive)
}

// GetUserReviews godoc
// @Summary      Получить PR пользователя
// @Description  Возвращает PR, где пользователь выступает ревьювером.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user_id  query     string  true  "Идентификатор пользователя"
// @Success      200      {object}  dto.UserReviewResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/users/getReview [get]
func (h *UserHandler) GetUserReviews(c *gin.Context) {
	log := logger(c)
	userID := c.Query("user_id")
	if userID == "" {
		log.Warnw("user_id query parameter missing")
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "user_id is required")
		return
	}
	log.Debugw("get user reviews request", "user_id", userID)

	prs, err := h.userSvc.GetUserReviews(userID)
	if err != nil {
		log.Errorw("failed to get user reviews", "user_id", userID, "error", err)
		h.handleDomainError(c, err)
		return
	}

	resp := mapper.BuildUserReviewResponse(userID, prs)
	c.JSON(http.StatusOK, resp)
	log.Infow("user reviews fetched", "user_id", userID, "pr_count", len(prs))
}

func (h *UserHandler) handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, serviceerrs.ErrUserNotFound):
		writeError(c, http.StatusNotFound, errorCodeNotFound, err.Error())
	case errors.Is(err, serviceerrs.ErrTeamNotFound):
		writeError(c, http.StatusNotFound, errorCodeNotFound, err.Error())
	default:
		writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
	}
}

// BulkDeactivate godoc
// @Summary      Массовая деактивация пользователей команды
// @Description  Деактивирует переданный список user_id внутри команды и безопасно переназначает их в открытых PR (если найдены кандидаты).
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      dto.BulkDeactivateRequest  true  "Команда и user_id"
// @Success      200      {object}  dto.BulkDeactivateResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/users/bulkDeactivate [post]
func (h *UserHandler) BulkDeactivate(c *gin.Context) {
	log := logger(c)
	var req dto.BulkDeactivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnw("invalid bulk deactivate payload", "error", err)
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}
	if req.TeamName == "" || len(req.UserIDs) == 0 {
		log.Warnw("missing fields in bulk deactivate", "payload", req)
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "team_name and user_ids are required")
		return
	}

	result, err := h.userSvc.BulkDeactivate(req.TeamName, req.UserIDs)
	if err != nil {
		log.Errorw("bulk deactivate failed", "team", req.TeamName, "error", err)
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.BulkDeactivateResponse{
		Team:        result.TeamName,
		Deactivated: result.DeactivatedUsers,
		Reassigned:  result.ReassignmentsDone,
		Skipped:     result.ReassignmentsSkipped,
		AffectedPRs: result.AffectedPullRequests,
	})
	log.Infow("bulk deactivate completed", "team", req.TeamName, "deactivated", result.DeactivatedUsers, "reassigned", result.ReassignmentsDone, "skipped", result.ReassignmentsSkipped, "prs", result.AffectedPullRequests)
}
