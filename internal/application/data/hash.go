package data

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// HashQueryRequest cria um hash estável do payload sem expor SQL.
func HashQueryRequest(req QueryRequest) (string, error) {
	// normalizar filtros para ordem determinística
	type filterItem struct {
		Key   string      `json:"key"`
		Value FilterField `json:"value"`
	}
	filters := make([]filterItem, 0, len(req.Filter))
	for k, v := range req.Filter {
		filters = append(filters, filterItem{Key: k, Value: v})
	}
	sort.Slice(filters, func(i, j int) bool { return filters[i].Key < filters[j].Key })

	canonical := struct {
		Schema     string       `json:"schema"`
		Limit      int          `json:"limit"`
		Offset     int          `json:"offset"`
		CountTotal bool         `json:"countTotal"`
		OrderBy    []OrderField `json:"orderBy"`
		Filter     []filterItem `json:"filter"`
		Fields     []string     `json:"fields"`
	}{
		Schema:     req.Schema,
		Limit:      req.Limit,
		Offset:     req.Offset,
		CountTotal: req.CountTotal,
		OrderBy:    req.OrderBy,
		Filter:     filters,
		Fields:     req.Fields,
	}

	b, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
