package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"api-database/internal/application/data"
)

// DataHandler expõe endpoints de dados.
type DataHandler struct {
	service *data.QueryService
}

func NewDataHandler(service *data.QueryService) *DataHandler {
	return &DataHandler{service: service}
}

func (h *DataHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")
	table := chi.URLParam(r, "table")

	var req data.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	resp, err := h.service.QueryTable(r.Context(), source, table, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
