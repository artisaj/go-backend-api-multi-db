package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect inicializa um client MongoDB e realiza um ping para validar a conexão.
// Em desenvolvimento, retorna nil sem erro se a conexão falhar (opcional).
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return client, nil
}

// ConnectOptional tenta conectar, mas não falha se a conexão não estiver disponível (útil em dev).
func ConnectOptional(ctx context.Context, uri string) *mongo.Client {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	client, err := mongo.Connect(pingCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil
	}

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil
	}

	return client
}
