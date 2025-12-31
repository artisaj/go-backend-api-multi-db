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

func (r *APIKeyRepositoryMongo) List(ctx context.Context) ([]apikey.APIKey, error) {
	col := r.client.Database(r.dbName).Collection(apiKeysCollection)
	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var keys []apikey.APIKey
	if err = cursor.All(ctx, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *APIKeyRepositoryMongo) Create(ctx context.Context, ak *apikey.APIKey) error {
	col := r.client.Database(r.dbName).Collection(apiKeysCollection)
	_, err := col.InsertOne(ctx, ak)
	return err
}

func (r *APIKeyRepositoryMongo) Update(ctx context.Context, key string, ak *apikey.APIKey) error {
	col := r.client.Database(r.dbName).Collection(apiKeysCollection)
	filter := bson.M{"key": key}
	update := bson.M{"$set": ak}
	_, err := col.UpdateOne(ctx, filter, update)
	return err
}

func (r *APIKeyRepositoryMongo) Delete(ctx context.Context, key string) error {
	col := r.client.Database(r.dbName).Collection(apiKeysCollection)
	filter := bson.M{"key": key}
	_, err := col.DeleteOne(ctx, filter)
	return err
}
