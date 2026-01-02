package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"api-database/internal/domain/job"
)

const jobsCollection = "jobs"

// JobRepositoryMongo implementa JobRepository usando MongoDB.
type JobRepositoryMongo struct {
	client *mongo.Client
	dbName string
}

func NewJobRepository(client *mongo.Client, dbName string) *JobRepositoryMongo {
	return &JobRepositoryMongo{client: client, dbName: dbName}
}

func (r *JobRepositoryMongo) collection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(jobsCollection)
}

func (r *JobRepositoryMongo) Insert(ctx context.Context, j *job.QueryJob) error {
	_, err := r.collection().InsertOne(ctx, j)
	return err
}

func (r *JobRepositoryMongo) UpdateStatus(ctx context.Context, id string, status string, fields map[string]any) error {
	update := bson.M{"$set": bson.M{"status": status}}
	if len(fields) > 0 {
		for k, v := range fields {
			update["$set"].(bson.M)[k] = v
		}
	}
	res, err := r.collection().UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("job not found")
	}
	return nil
}

func (r *JobRepositoryMongo) GetByID(ctx context.Context, id string) (*job.QueryJob, error) {
	var j job.QueryJob
	if err := r.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&j); err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *JobRepositoryMongo) GetByPayloadHash(ctx context.Context, payloadHash string) ([]*job.QueryJob, error) {
	cursor, err := r.collection().Find(ctx, bson.M{"payloadHash": payloadHash})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []*job.QueryJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}
