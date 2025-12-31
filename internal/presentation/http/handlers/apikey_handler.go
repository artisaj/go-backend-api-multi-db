package handlers

import (
	"net/http"
	"time"

	"api-database/internal/domain/apikey"
	"api-database/internal/presentation/http/middleware"
)

type APIKeyHandler struct {
	repo apikey.APIKeyRepository
}

func NewAPIKeyHandler(repo apikey.APIKeyRepository) *APIKeyHandler {
	return &APIKeyHandler{repo: repo}
}

// GetMe retorna a chave do usuário autenticado
func (h *APIKeyHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ak := middleware.GetAPIKeyFromContext(r.Context())
	if ak == nil {
		http.Error(w, `{"code":"NO_API_KEY","message":"no API key provided"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	respondJSON(w, ak)
}

// ListKeys retorna todas as chaves (sem exposição das chaves)
func (h *APIKeyHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, `{"code":"ERROR","message":"failed to list keys"}`, http.StatusInternalServerError)
		return
	}

	// Ocultar o valor real das chaves na resposta
	var safeKeys []map[string]interface{}
	for _, k := range keys {
		safeKeys = append(safeKeys, map[string]interface{}{
			"key":         k.Key[:8] + "...",
			"name":        k.Name,
			"description": k.Description,
			"createdAt":   k.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	respondJSON(w, safeKeys)
}

// CreateKey cria uma nova chave
func (h *APIKeyHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string                `json:"name"`
		Description string                `json:"description"`
		Permissions []apikey.Permission `json:"permissions"`
	}

	if err := parseJSONBody(r, &req); err != nil {
		http.Error(w, `{"code":"INVALID_JSON","message":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"code":"INVALID_NAME","message":"name is required"}`, http.StatusBadRequest)
		return
	}

	newKey := &apikey.APIKey{
		Key:         apikey.GenerateKey(),
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.Create(r.Context(), newKey); err != nil {
		http.Error(w, `{"code":"ERROR","message":"failed to create key"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	respondJSON(w, newKey)
}

// UpdateKey atualiza a descrição de uma chave
func (h *APIKeyHandler) UpdateKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, `{"code":"INVALID_KEY","message":"key is required"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := parseJSONBody(r, &req); err != nil {
		http.Error(w, `{"code":"INVALID_JSON","message":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	existingKey, err := h.repo.GetByKey(r.Context(), key)
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"key not found"}`, http.StatusNotFound)
		return
	}

	existingKey.Name = req.Name
	existingKey.Description = req.Description
	existingKey.UpdatedAt = time.Now()

	if err := h.repo.Update(r.Context(), key, existingKey); err != nil {
		http.Error(w, `{"code":"ERROR","message":"failed to update key"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	respondJSON(w, existingKey)
}

// DeleteKey deleta uma chave
func (h *APIKeyHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, `{"code":"INVALID_KEY","message":"key is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), key); err != nil {
		http.Error(w, `{"code":"ERROR","message":"failed to delete key"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
