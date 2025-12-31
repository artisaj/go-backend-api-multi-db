package datasource

import "context"

// DataSource representa a configuração de uma fonte de dados.
type DataSource struct {
	Name         string         `bson:"name" json:"name"`
	Type         string         `bson:"type" json:"type"`
	Description  string         `bson:"description" json:"description"`
	Connection   Connection     `bson:"connection" json:"connection"`
	Capabilities Capabilities   `bson:"capabilities" json:"capabilities"`
	Limits       Limits         `bson:"limits" json:"limits"`
	Version      int            `bson:"version" json:"version"`
	CreatedAt    interface{}    `bson:"createdAt" json:"createdAt"`
	UpdatedAt    interface{}    `bson:"updatedAt" json:"updatedAt"`
	Raw          map[string]any `bson:"-" json:"-"`
	Extra        map[string]any `bson:",inline" json:"-"`
}

// Connection contém parâmetros de conexão.
type Connection struct {
	Host     string `bson:"host" json:"host"`
	Port     int    `bson:"port" json:"port"`
	User     string `bson:"user" json:"user"`
	Password string `bson:"password" json:"password"`
	Database string `bson:"database" json:"database"`
	SSLMode  string `bson:"sslMode" json:"sslMode"`
}

// Capabilities descreve limites de recursos.
type Capabilities struct {
	SupportsJoins        bool `bson:"supportsJoins" json:"supportsJoins"`
	SupportsTransactions bool `bson:"supportsTransactions" json:"supportsTransactions"`
	MaxDepthLimit        int  `bson:"maxDepthLimit" json:"maxDepthLimit"`
}

// Limits define restrições de execução.
type Limits struct {
	MaxRows        int `bson:"maxRows" json:"maxRows"`
	QueryTimeoutMs int `bson:"queryTimeoutMs" json:"queryTimeoutMs"`
}

// DataSourceRepository interface para buscar configurações.
type DataSourceRepository interface {
	GetByName(ctx context.Context, name string) (*DataSource, error)
}
