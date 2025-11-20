package dto

// @Description Стандартный ответ с ошибкой.
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Детали ошибки.
	Error ErrorBody `json:"error" validate:"required"`
} // @name ErrorResponse

// @Description Содержит код и сообщение ошибки.
// swagger:model ErrorBody
type ErrorBody struct {
	// Код ошибки.
	Code string `json:"code" validate:"required" example:"NOT_FOUND"`
	// Сообщение ошибки.
	Message string `json:"message" validate:"required" example:"resource not found"`
} // @name ErrorBody
