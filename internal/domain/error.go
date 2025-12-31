package domain

import "net/http"

// ErrorCode representa um código de erro padronizado.
type ErrorCode string

const (
	ErrInvalidInput       ErrorCode = "INVALID_INPUT"
	ErrInvalidTable       ErrorCode = "INVALID_TABLE"
	ErrInvalidSchema      ErrorCode = "INVALID_SCHEMA"
	ErrColumnBlocked      ErrorCode = "COLUMN_BLOCKED"
	ErrDataSourceNotFound ErrorCode = "DATASOURCE_NOT_FOUND"
	ErrUnsupportedType    ErrorCode = "UNSUPPORTED_TYPE"
	ErrQueryFailed        ErrorCode = "QUERY_FAILED"
	ErrInternal           ErrorCode = "INTERNAL_ERROR"
)

// AppError representa um erro estruturado da aplicação.
type AppError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	status  int
}

// NewAppError cria um novo erro estruturado.
func NewAppError(code ErrorCode, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: nil,
		status:  status,
	}
}

// WithDetails adiciona detalhes ao erro.
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// Status retorna o código HTTP apropriado.
func (e *AppError) Status() int {
	return e.status
}

// Error implementa interface error.
func (e *AppError) Error() string {
	return e.Message
}

// MapErrorToHTTP mapeia erros de negócio para status HTTP.
func MapErrorToHTTP(err error) (int, *AppError) {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Status(), appErr
	}

	// Erro genérico não tipado = 500
	return http.StatusInternalServerError, NewAppError(
		ErrInternal,
		"Internal server error",
		http.StatusInternalServerError,
	)
}
