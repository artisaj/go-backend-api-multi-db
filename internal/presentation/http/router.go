package httpserver

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"api-database/internal/config"
	httpmiddleware "api-database/internal/presentation/http/middleware"
)

// NewRouter configura middlewares base e rotas públicas.
func NewRouter(cfg config.Config, logger zerolog.Logger, dataHandler *DataHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(cfg.Thresholds.QueryTimeoutMs) * time.Millisecond))
	r.Use(httpmiddleware.Logging(logger))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","env":"` + cfg.Env + `"}`))
	})

	if dataHandler != nil {
		r.Post("/data/{source}/{table}", dataHandler.HandleQuery)
	}

	return r
}
