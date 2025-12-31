package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"api-database/internal/domain/datasource"
)

const dataSourcesCollection = "data_sources"

// DataSourceRepositoryMongo implementa DataSourceRepository usando MongoDB.
type DataSourceRepositoryMongo struct {
	client *mongo.Client
	dbName string
}

func NewDataSourceRepository(client *mongo.Client, dbName string) *DataSourceRepositoryMongo {
	return &DataSourceRepositoryMongo{client: client, dbName: dbName}
}

func (r *DataSourceRepositoryMongo) GetByName(ctx context.Context, name string) (*datasource.DataSource, error) {
	col := r.client.Database(r.dbName).Collection(dataSourcesCollection)
	filter := bson.M{"name": name}
	opts := options.FindOne()

	var ds datasource.DataSource
	if err := col.FindOne(ctx, filter, opts).Decode(&ds); err != nil {
		return nil, err
	}
	return &ds, nil
}
