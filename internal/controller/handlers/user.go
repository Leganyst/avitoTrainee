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
	var req dto.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid request payload")
		return
	}

	user, err := h.userSvc.SetActive(req.UserID, req.IsActive)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		User: mapper.MapUserToDTO(*user),
	})
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
	userID := c.Query("user_id")
	if userID == "" {
		writeError(c, http.StatusBadRequest, errorCodeBadRequest, "user_id is required")
		return
	}

	prs, err := h.userSvc.GetUserReviews(userID)
	if err != nil {
		h.handleDomainError(c, err)
		return
	}

	resp := mapper.BuildUserReviewResponse(userID, prs)
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, serviceerrs.ErrUserNotFound):
		writeError(c, http.StatusNotFound, errorCodeNotFound, err.Error())
	default:
		writeError(c, http.StatusInternalServerError, errorCodeInternal, "internal error")
	}
}
