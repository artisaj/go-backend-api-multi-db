package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"api-database/internal/application/data"
	"api-database/internal/domain"
	"api-database/internal/telemetry"
)

// DataHandler expõe endpoints de dados.
type DataHandler struct {
	service *data.QueryService
	metrics *telemetry.Metrics
}

func NewDataHandler(service *data.QueryService, metrics *telemetry.Metrics) *DataHandler {
	return &DataHandler{service: service, metrics: metrics}
}

func (h *DataHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")
	table := chi.URLParam(r, "table")

	var req data.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := domain.NewAppError(domain.ErrInvalidInput, "invalid request body", http.StatusBadRequest)
		writeError(w, appErr)
		return
	}

	resp, err := h.service.QueryTable(r.Context(), source, table, req)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if !ok {
			appErr = domain.NewAppError(domain.ErrInternal, err.Error(), http.StatusInternalServerError)
		}
		writeError(w, appErr)

		// Registrar métrica de erro
		if h.metrics != nil {
			h.metrics.RecordQuery(telemetry.QueryMetric{
				DataSource: source,
				Table:      table,
				Status:     "error",
				Latency:    0,
				Rows:       0,
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)

	// Registrar métrica de sucesso
	if h.metrics != nil {
		h.metrics.RecordQuery(telemetry.QueryMetric{
			DataSource: source,
			Table:      table,
			Status:     "success",
			Latency:    resp.Metadata.TookMs,
			Rows:       resp.Metadata.Rows,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, appErr *domain.AppError) {
	status := appErr.Status()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(appErr)
}
