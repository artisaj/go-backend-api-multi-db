package domain

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	err := NewAppError(ErrInvalidTable, "invalid table name", http.StatusBadRequest)
	
	assert.Equal(t, ErrInvalidTable, err.Code)
	assert.Equal(t, "invalid table name", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.Status())
	assert.Equal(t, "invalid table name", err.Error())
}

func TestAppErrorWithDetails(t *testing.T) {
	err := NewAppError(ErrColumnBlocked, "column blocked", http.StatusForbidden).
		WithDetails(map[string]interface{}{
			"column": "passwordHash",
			"table":  "User",
		})
	
	assert.NotNil(t, err.Details)
	assert.Equal(t, "passwordHash", err.Details["column"])
}

func TestMapErrorToHTTP(t *testing.T) {
	appErr := NewAppError(ErrDataSourceNotFound, "not found", http.StatusNotFound)
	status, mappedErr := MapErrorToHTTP(appErr)
	
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, appErr, mappedErr)
}

func TestMapErrorToHTTP_GenericError(t *testing.T) {
	genericErr := assert.AnError
	status, mappedErr := MapErrorToHTTP(genericErr)
	
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, ErrInternal, mappedErr.Code)
}
