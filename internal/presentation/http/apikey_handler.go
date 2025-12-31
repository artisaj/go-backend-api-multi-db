package httpserver

import (
	"encoding/json"
	"net/http"

	"api-database/internal/presentation/http/middleware"
)

// APIKeyHandler expõe endpoints de chaves de API.
type APIKeyHandler struct{}

func NewAPIKeyHandler() *APIKeyHandler {
	return &APIKeyHandler{}
}

func (h *APIKeyHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	ak := middleware.GetAPIKeyFromContext(r.Context())
	if ak == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "NO_API_KEY",
			"message": "no API key provided",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ak)
}
