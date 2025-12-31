package apikey

import (
	"context"

	"github.com/google/uuid"
)

// Permission define acesso a um recurso em um nível específico.
type Permission struct {
	// Resource examples: "racehub", "racehub.User", "racehub.User.passwordHash"
	Resource string `bson:"resource" json:"resource"`
	// Level: "database", "table", "column"
	Level string `bson:"level" json:"level"`
}

// APIKey representa uma chave de acesso.
type APIKey struct {
	Key         string       `bson:"key" json:"key"`
	Name        string       `bson:"name" json:"name"`
	Description string       `bson:"description" json:"description"`
	Permissions []Permission `bson:"permissions" json:"permissions"`
	CreatedAt   interface{}  `bson:"createdAt" json:"createdAt"`
	UpdatedAt   interface{}  `bson:"updatedAt" json:"updatedAt"`
}

// GenerateKey cria uma nova chave UUID v4.
func GenerateKey() string {
	return uuid.New().String()
}

// HasPermission verifica se a chave tem acesso a um recurso específico.
// Verifica em order: resource específico → table → database
func (ak *APIKey) HasPermission(resource string) bool {
	for _, p := range ak.Permissions {
		if p.Resource == resource {
			return true
		}
	}
	return false
}

// APIKeyRepository interface para gerenciar chaves.
type APIKeyRepository interface {
	GetByKey(ctx context.Context, key string) (*APIKey, error)
	List(ctx context.Context) ([]APIKey, error)
	Create(ctx context.Context, ak *APIKey) error
	Update(ctx context.Context, key string, ak *APIKey) error
	Delete(ctx context.Context, key string) error
}
