package postgres

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connector executa consultas simples em PostgreSQL.
type Connector struct {
	pool *pgxpool.Pool
}

// NewConnector cria um pool pgx a partir dos dados de conexão.
func NewConnector(ctx context.Context, host string, port int, user, password, database, sslMode string) (*Connector, error) {
	if sslMode == "" {
		sslMode = "disable"
	}
	pgURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/" + database,
		RawQuery: url.Values{
			"sslmode": []string{sslMode},
		}.Encode(),
	}
	connStr := pgURL.String()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Connector{pool: pool}, nil
}

// Query retorna linhas como slice de map[string]any.
func (c *Connector) Query(ctx context.Context, sql string, args ...any) ([]map[string]any, error) {
	rows, err := c.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	cols := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		cols[i] = string(fd.Name)
	}

	var result []map[string]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		item := make(map[string]any, len(cols))
		for i, col := range cols {
			item[col] = normalizeValue(values[i])
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// Close fecha o pool.
func (c *Connector) Close() {
	c.pool.Close()
}

// normalizeValue converte tipos binários/UUID/time em formas mais amigáveis.
func normalizeValue(v any) any {
	switch val := v.(type) {
	case [16]byte:
		return formatUUIDBytes(val[:])
	case []byte:
		if len(val) == 16 {
			return formatUUIDBytes(val)
		}
		return string(val)
	case time.Time:
		return val.UTC().Format(time.RFC3339Nano)
	default:
		return v
	}
}

// formatUUIDBytes converte 16 bytes em string UUID canonical (8-4-4-4-12).
func formatUUIDBytes(b []byte) string {
	if len(b) != 16 {
		return string(b)
	}
	hexStr := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexStr[0:8], hexStr[8:12], hexStr[12:16], hexStr[16:20], hexStr[20:32])
}
