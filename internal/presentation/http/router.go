package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"api-database/internal/config"
	"api-database/internal/domain/apikey"
	"api-database/internal/domain/datasource"
	httpmiddleware "api-database/internal/presentation/http/middleware"
	"api-database/internal/presentation/http/handlers"
	"api-database/internal/telemetry"
)

// NewRouter configura middlewares base e rotas públicas.
func NewRouter(cfg config.Config, logger zerolog.Logger, dataHandler *DataHandler, dsRepo datasource.DataSourceRepository, metrics *telemetry.Metrics, akRepo apikey.APIKeyRepository) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(cfg.Thresholds.QueryTimeoutMs) * time.Millisecond))
	r.Use(httpmiddleware.Logging(logger))

	// Auth middleware (opcional: requer X-API-Key header)
	if akRepo != nil {
		r.Use(httpmiddleware.AuthMiddleware(akRepo))
	}

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","env":"` + cfg.Env + `"}`))
	})

	// Métricas endpoint
	if metrics != nil {
		r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"summary": metrics.GetSummary(),
				"total":   len(metrics.GetMetrics()),
			})
		})
	}

	// Datasources endpoint
	if dsRepo != nil {
		r.Get("/datasources", func(w http.ResponseWriter, r *http.Request) {
			sources, err := dsRepo.ListAll(r.Context())
			w.Header().Set("Content-Type", "application/json")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "failed to list datasources"})
				return
			}
			// Retornar apenas name, type, description (não exposar connection details)
			type SafeDS struct {
				Name        string `json:"name"`
				Type        string `json:"type"`
				Description string `json:"description"`
			}
			safe := make([]SafeDS, len(sources))
			for i, s := range sources {
				safe[i] = SafeDS{Name: s.Name, Type: s.Type, Description: s.Description}
			}
			json.NewEncoder(w).Encode(safe)
		})
	}

	if dataHandler != nil {
		r.Post("/data/{source}/{table}", dataHandler.HandleQuery)
	}

	// API Key CRUD endpoints
	if akRepo != nil {
		akHandler := handlers.NewAPIKeyHandler(akRepo)
		r.Get("/api-keys/me", akHandler.GetMe)
		r.Get("/api-keys", akHandler.ListKeys)
		r.Post("/api-keys", akHandler.CreateKey)
		r.Put("/api-keys/{key}", akHandler.UpdateKey)
		r.Delete("/api-keys/{key}", akHandler.DeleteKey)
	}

	// Servir dashboard
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/index.html")
	})

	return r
}
