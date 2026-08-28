package httpapi

import (
	"errors"
	"log"
	"net/http"

	"checkout-api/internal/repository"
	"checkout-api/internal/service"
	"checkout-api/internal/validation"
)

const (
	CodeValidationFailed  = "VALIDATION_FAILED"
	CodeInvalidRequest    = "INVALID_REQUEST"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeNotFound          = "NOT_FOUND"
	CodeConflict          = "CONFLICT"
	CodeInsufficientStock = "INSUFFICIENT_STOCK"
	CodePaymentDeclined   = "PAYMENT_DECLINED"
	CodeInternal          = "INTERNAL_ERROR"
)

type ErrorEnvelope struct {
	Error   string                  `json:"error"`
	Message string                  `json:"message"`
	Details []validation.FieldError `json:"details,omitempty"`
}

var defaultMessages = map[string]string{
	CodeValidationFailed:  "validation failed",
	CodeInvalidRequest:    "invalid request",
	CodeUnauthorized:      "authentication required",
	CodeForbidden:         "forbidden",
	CodeNotFound:          "resource not found",
	CodeConflict:          "resource already exists",
	CodeInsufficientStock: "insufficient stock",
	CodePaymentDeclined:   "payment declined",
	CodeInternal:          "internal server error",
}

func Error(w http.ResponseWriter, err error) {
	status, env := classify(err)
	if status == http.StatusInternalServerError {
		log.Printf("internal error: %v", err)
	}
	writeJSON(w, status, env)
}

func BadRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, ErrorEnvelope{
		Error:   CodeInvalidRequest,
		Message: msg,
	})
}

func classify(err error) (int, ErrorEnvelope) {
	var valErr *validation.ValidationError
	switch {
	case errors.As(err, &valErr):
		return http.StatusUnprocessableEntity, ErrorEnvelope{
			Error:   CodeValidationFailed,
			Message: defaultMessages[CodeValidationFailed],
			Details: valErr.Fields,
		}
	case errors.Is(err, repository.ErrNotFound):
		return http.StatusNotFound, envelope(CodeNotFound)
	case errors.Is(err, repository.ErrConflict):
		return http.StatusConflict, envelope(CodeConflict)
	case errors.Is(err, repository.ErrForbidden):
		return http.StatusForbidden, envelope(CodeForbidden)
	case errors.Is(err, repository.ErrInsufficient):
		return http.StatusConflict, envelope(CodeInsufficientStock)
	case errors.Is(err, service.ErrInvalidCredentials):
		return http.StatusUnauthorized, envelope(CodeUnauthorized)
	case errors.Is(err, service.ErrPaymentDeclined):
		return http.StatusPaymentRequired, envelope(CodePaymentDeclined)
	default:
		return http.StatusInternalServerError, envelope(CodeInternal)
	}
}

func envelope(code string) ErrorEnvelope {
	return ErrorEnvelope{Error: code, Message: defaultMessages[code]}
}
