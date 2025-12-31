# Tutorial: Registrar PostgreSQL como Datasource

## Objetivo
Integrar a base de dados PostgreSQL `postgresql://user:password@localhost:5432/racehub` ao centralizador de dados via API.

## Fluxo geral

```
1. Inserir documento em MongoDB (data_sources)
   ↓
2. API carrega config ao iniciar ou sob demanda
   ↓
3. Factory de conectores cria instância do Postgres
   ↓
4. Requisições POST /data/racehub/{tabela} rotineiam para o conector Postgres
```

## Passo 1: Registrar datasource no MongoDB

Via MongoDB Compass, na coleção `data_sources`, inserir (JSON válido sem `new Date()`):

```json
{
  "name": "racehub",
  "type": "postgres",
  "description": "Base de dados de corridas",
  "connection": {
    "host": "localhost",
    "port": 5432,
    "user": "user",
    "password": "password",
    "database": "racehub",
    "sslMode": "disable"
  },
  "capabilities": {
    "supportsJoins": true,
    "supportsTransactions": true,
    "maxDepthLimit": 5
  },
  "limits": {
    "maxRows": 10000,
    "queryTimeoutMs": 30000
  },
  "version": 1,
  "createdAt": "2025-12-31T00:00:00.000Z",
  "updatedAt": "2025-12-31T00:00:00.000Z"
}
```

No Compass:
- Abra `api_database_config` → `data_sources` → **Insert Document** → cole o JSON acima e salve.
- Se preferir, remova `createdAt`/`updatedAt` e deixe o sistema preencher depois.

## Passo 2: Fazer uma requisição de dados

Após implementar o endpoint `POST /data/{fonte}/{tabela}`:

```bash
curl -X POST http://localhost:8080/data/racehub/usuarios \
  -H "Content-Type: application/json" \
  -d '{
    "filter": {
      "id": { "$gt": 10 }
    },
    "orderBy": [
      { "field": "created_at", "direction": "desc" }
    ],
    "limit": 100,
    "offset": 0
  }'
```

## Passo 3: Entender a resposta (sync vs async)

### Resposta rápida (sync):
```json
{
  "data": [
    { "id": 100, "name": "João", "created_at": "2025-01-01T10:00:00Z" },
    { "id": 101, "name": "Maria", "created_at": "2025-01-01T11:00:00Z" }
  ],
  "metadata": {
    "total": 2,
    "took_ms": 45,
    "cached": false
  }
}
```

### Resposta lenta (auto-async via query-id):
```json
{
  "queryId": "q_8f9c2d1e5a7b",
  "status": "processing",
  "message": "Query will take longer; check status via GET /queries/q_8f9c2d1e5a7b"
}
```

Cliente então faz polling:
```bash
curl http://localhost:8080/queries/q_8f9c2d1e5a7b
```

Quando pronto:
```json
{
  "queryId": "q_8f9c2d1e5a7b",
  "status": "completed",
  "data": [ ... ], // dados completos
  "metadata": {
    "total": 50000,
    "took_ms": 12500,
    "cached": false
  }
}
```

## Passo 4: Implementação interna (roadmap)

Os passos a seguir devem ser implementados:

1. **Carregar DataSource do MongoDB** → `DataSourceRepository`
2. **Factory de Conectores** → Criar conector PostgreSQL baseado no tipo
3. **DTOs e Validação** → Request/Response structs com constraints (depth-limit, etc.)
4. **Calculadora de Custo** → Estimar tempo via `query_metrics` histórico
5. **Switch Sync/Async** → Se custo > threshold → enfileirar em `jobs`
6. **Executor de Queries** → Traduzir filtros genéricos para SQL/Mongo
7. **Cache** → query-hash + entity-version; TTL + invalidação por eventos

## Estrutura esperada de `query_metrics`

Usado para estimar custos:

```json
{
  "dataSourceId": ObjectId("..."),
  "table": "usuarios",
  "queryShape": "filter:id;orderBy:created_at",
  "p50Ms": 120,
  "p95Ms": 450,
  "p99Ms": 890,
  "timeoutCount": 2,
  "averageRowsReturned": 250,
  "averagePayloadBytes": 15000,
  "version": 1,
  "lastUpdated": ISODate("2025-12-31T17:00:00Z")
}
```

Se p95Ms > ASYNC_SWITCH_P95_MS (3500ms), a query rota para async automaticamente.

## Próximos passos de código

1. [ ] Interface `DatabaseConnector` (Query, Execute, Transaction)
2. [ ] Implementação `PostgresConnector` (pgx driver)
3. [ ] `DataSourceRepository` (MongoDB)
4. [ ] `QueryMetricsRepository` (MongoDB)
5. [ ] `ConnectorFactory` (seleção by type)
6. [ ] DTO `DataQueryRequest` e `QueryResponse`
7. [ ] `CostEstimator` (p95 + heurística)
8. [ ] Router que detecta sync vs async
9. [ ] Job Queue (`JobRepository`, worker background)
10. [ ] Cache com query-hash + entity-version
