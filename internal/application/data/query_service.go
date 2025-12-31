package data

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"api-database/internal/domain"
	"api-database/internal/domain/datasource"
	"api-database/internal/infrastructure/postgres"
)

// QueryService executa consultas simples em uma tabela de um datasource.
type QueryService struct {
	repo datasource.DataSourceRepository
}

func NewQueryService(repo datasource.DataSourceRepository) *QueryService {
	return &QueryService{repo: repo}
}

// QueryRequest define entrada mínima para teste inicial.
type QueryRequest struct {
	Schema     string                 `json:"schema"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	CountTotal bool                   `json:"countTotal"`
	OrderBy    []OrderField           `json:"orderBy"`
	Filter     map[string]FilterField `json:"filter"`
	Fields     []string               `json:"fields"` // Colunas a retornar; se vazio, SELECT *
}

// QueryResponse retorna dados e metadados simples.
type QueryResponse struct {
	Data     []map[string]any `json:"data"`
	Metadata Meta             `json:"metadata"`
}

// Meta inclui info básica.
type Meta struct {
	Rows   int    `json:"rows"`
	Table  string `json:"table"`
	TookMs int64  `json:"tookMs"`
	Total  *int64 `json:"total,omitempty"`
}

// OrderField define campo e direção.
type OrderField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// FilterField suporta comparações simples.
type FilterField struct {
	Eq  any `json:"$eq,omitempty"`
	Gt  any `json:"$gt,omitempty"`
	Gte any `json:"$gte,omitempty"`
	Lt  any `json:"$lt,omitempty"`
	Lte any `json:"$lte,omitempty"`
}

var tableNameRegex = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
var directionRegex = regexp.MustCompile(`^(?i)(asc|desc)$`)
var columnRegex = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

const defaultMaxLimit = 500

func (s *QueryService) QueryTable(ctx context.Context, sourceName, table string, req QueryRequest) (*QueryResponse, error) {
	if !tableNameRegex.MatchString(table) {
		return nil, domain.NewAppError(domain.ErrInvalidTable, "invalid table name", http.StatusBadRequest)
	}
	if req.Schema != "" && !tableNameRegex.MatchString(req.Schema) {
		return nil, domain.NewAppError(domain.ErrInvalidSchema, "invalid schema name", http.StatusBadRequest)
	}

	ds, err := s.repo.GetByName(ctx, sourceName)
	if err != nil {
		return nil, domain.NewAppError(domain.ErrDataSourceNotFound, "datasource not found", http.StatusNotFound)
	}
	if ds.Type != "postgres" {
		return nil, domain.NewAppError(domain.ErrUnsupportedType, "only postgres is supported", http.StatusBadRequest)
	}

	// Validar colunas bloqueadas em filter
	for col := range req.Filter {
		if isColumnBlocked(table, col, ds.BlockedColumns) {
			return nil, domain.NewAppError(domain.ErrColumnBlocked, fmt.Sprintf("column blocked: %s", col), http.StatusForbidden)
		}
	}

	// Validar colunas bloqueadas em orderBy
	for _, o := range req.OrderBy {
		if isColumnBlocked(table, o.Field, ds.BlockedColumns) {
			return nil, domain.NewAppError(domain.ErrColumnBlocked, fmt.Sprintf("column blocked: %s", o.Field), http.StatusForbidden)
		}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	maxLimit := defaultMaxLimit
	if ds.Limits.MaxRows > 0 && ds.Limits.MaxRows < maxLimit {
		maxLimit = ds.Limits.MaxRows
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Validar fields contra blockedColumns
	for _, field := range req.Fields {
		if !columnRegex.MatchString(field) {
			return nil, domain.NewAppError(domain.ErrInvalidInput, fmt.Sprintf("invalid column name: %s", field), http.StatusBadRequest)
		}
		if isColumnBlocked(table, field, ds.BlockedColumns) {
			return nil, domain.NewAppError(domain.ErrColumnBlocked, fmt.Sprintf("column blocked: %s", field), http.StatusForbidden)
		}
	}

	conn, err := postgres.NewConnector(ctx, ds.Connection.Host, ds.Connection.Port, ds.Connection.User, ds.Connection.Password, ds.Connection.Database, ds.Connection.SSLMode)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	fullTable := quoteIdent(table)
	if req.Schema != "" {
		fullTable = fmt.Sprintf("%s.%s", quoteIdent(req.Schema), quoteIdent(table))
	}

	// Construir SELECT com fields ou *
	var selectClause string
	if len(req.Fields) > 0 {
		cols := make([]string, len(req.Fields))
		for i, f := range req.Fields {
			cols[i] = quoteIdent(f)
		}
		selectClause = strings.Join(cols, ", ")
	} else {
		selectClause = "*"
	}

	whereClause, params := buildWhere(req.Filter)
	orderClause := buildOrder(req.OrderBy)

	paramIdx := len(params) + 1
	params = append(params, limit)
	params = append(params, offset)

	query := fmt.Sprintf("SELECT %s FROM %s %s %s LIMIT $%d OFFSET $%d", selectClause, fullTable, whereClause, orderClause, paramIdx, paramIdx+1)

	start := time.Now()
	rows, err := conn.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	took := time.Since(start).Milliseconds()

	// Remover colunas bloqueadas da resposta
	for i := range rows {
		for _, blocked := range ds.BlockedColumns {
			parts := strings.SplitN(blocked, ".", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], table) {
				delete(rows[i], parts[1])
			}
		}
	}

	var totalPtr *int64
	if req.CountTotal {
		countQuery := fmt.Sprintf("SELECT COUNT(1) AS total FROM %s %s", fullTable, whereClause)
		totalRows, err := conn.Query(ctx, countQuery, params[:len(params)-2]...)
		if err == nil && len(totalRows) > 0 {
			if v, ok := totalRows[0]["total"].(int64); ok {
				totalPtr = &v
			}
		}
	}

	return &QueryResponse{
		Data: rows,
		Metadata: Meta{
			Rows:   len(rows),
			Table:  table,
			TookMs: took,
			Total:  totalPtr,
		},
	}, nil
}

func quoteIdent(name string) string {
	return "\"" + strings.ReplaceAll(name, "\"", "\"\"") + "\""
}

func buildWhere(filters map[string]FilterField) (string, []any) {
	if len(filters) == 0 {
		return "", nil
	}
	var clauses []string
	var params []any
	idx := 1
	for col, f := range filters {
		if !columnRegex.MatchString(col) {
			continue
		}
		colIdent := quoteIdent(col)
		if f.Eq != nil {
			clauses = append(clauses, fmt.Sprintf("%s = $%d", colIdent, idx))
			params = append(params, f.Eq)
			idx++
		}
		if f.Gt != nil {
			clauses = append(clauses, fmt.Sprintf("%s > $%d", colIdent, idx))
			params = append(params, f.Gt)
			idx++
		}
		if f.Gte != nil {
			clauses = append(clauses, fmt.Sprintf("%s >= $%d", colIdent, idx))
			params = append(params, f.Gte)
			idx++
		}
		if f.Lt != nil {
			clauses = append(clauses, fmt.Sprintf("%s < $%d", colIdent, idx))
			params = append(params, f.Lt)
			idx++
		}
		if f.Lte != nil {
			clauses = append(clauses, fmt.Sprintf("%s <= $%d", colIdent, idx))
			params = append(params, f.Lte)
			idx++
		}
	}
	if len(clauses) == 0 {
		return "", params
	}
	return "WHERE " + strings.Join(clauses, " AND "), params
}

func buildOrder(order []OrderField) string {
	if len(order) == 0 {
		return ""
	}
	var clauses []string
	for _, o := range order {
		if !columnRegex.MatchString(o.Field) {
			continue
		}
		dir := "ASC"
		if directionRegex.MatchString(o.Direction) {
			dir = strings.ToUpper(o.Direction)
		}
		clauses = append(clauses, fmt.Sprintf("%s %s", quoteIdent(o.Field), dir))
	}
	if len(clauses) == 0 {
		return ""
	}
	return "ORDER BY " + strings.Join(clauses, ", ")
}

// isColumnBlocked verifica se table.column está na lista de bloqueados
func isColumnBlocked(table, column string, blocked []string) bool {
	fullCol := strings.ToLower(fmt.Sprintf("%s.%s", table, column))
	for _, b := range blocked {
		if strings.ToLower(b) == fullCol {
			return true
		}
	}
	return false
}
