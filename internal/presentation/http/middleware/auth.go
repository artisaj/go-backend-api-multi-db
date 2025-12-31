package middleware

import (
	"context"
	"net/http"

	"api-database/internal/domain/apikey"
)

const contextKeyAPIKey = "api_key"

// AuthMiddleware valida X-API-Key header e anexa a chave ao contexto.
func AuthMiddleware(akRepo apikey.APIKeyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				// Se não houver header, deixar passar (opcional: pode rejeitar com 401)
				next.ServeHTTP(w, r)
				return
			}

			ak, err := akRepo.GetByKey(r.Context(), key)
			if err != nil || ak == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code":"INVALID_API_KEY","message":"invalid or expired API key"}`))
				return
			}

			// Anexar chave ao contexto
			ctx := context.WithValue(r.Context(), contextKeyAPIKey, ak)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAPIKeyFromContext extrai a chave do contexto.
func GetAPIKeyFromContext(ctx context.Context) *apikey.APIKey {
	ak, ok := ctx.Value(contextKeyAPIKey).(*apikey.APIKey)
	if !ok {
		return nil
	}
	return ak
}
