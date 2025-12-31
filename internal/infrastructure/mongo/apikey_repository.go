package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"api-database/internal/domain/apikey"
)

const apiKeysCollection = "api_keys"

// APIKeyRepositoryMongo implementa APIKeyRepository usando MongoDB.
type APIKeyRepositoryMongo struct {
	client *mongo.Client
	dbName string
}

func NewAPIKeyRepository(client *mongo.Client, dbName string) *APIKeyRepositoryMongo {
	return &APIKeyRepositoryMongo{client: client, dbName: dbName}
}

func (r *APIKeyRepositoryMongo) GetByKey(ctx context.Context, key string) (*apikey.APIKey, error) {
	col := r.client.Database(r.dbName).Collection(apiKeysCollection)
	filter := bson.M{"key": key}

	var ak apikey.APIKey
	if err := col.FindOne(ctx, filter).Decode(&ak); err != nil {
		return nil, err
	}
	return &ak, nil
}
