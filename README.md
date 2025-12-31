# API Database

API HTTP simples que expõe consultas read-only em Postgres a partir de configurações guardadas no MongoDB. Útil para centralizar acesso a dados de múltiplas fontes com validação básica de filtros, ordenação e limites.

## Requisitos
- Go 1.22+
- Docker e Docker Compose
- MongoDB (para armazenar configurações de datasources)
- Postgres acessível para o datasource configurado

## Configuração rápida
1. Opcional: copie `.env.example` para `.env` e ajuste variáveis (porta, log, URIs).
2. Suba os serviços com Docker Compose:

   ```sh
   docker-compose up -d
   ```

   - Sobe MongoDB e a API em `localhost:8080`.
   - Mongo é iniciado com usuário `admin/admin123` e banco `api_database_config`.

3. Cadastre um datasource na coleção `data_sources` do Mongo (no banco `api_database_config`). Exemplo para Postgres na máquina host:

   ```js
   db.data_sources.insertOne({
     name: "main",
     type: "postgres",
     description: "Postgres principal",
     connection: {
       host: "host.docker.internal",
       port: 5432,
       user: "postgres",
       password: "postgres",
       database: "postgres",
       sslMode: "disable"
     },
     capabilities: {
       supportsJoins: true,
       supportsTransactions: true,
       maxDepthLimit: 0
     },
     limits: {
       maxRows: 500,
       queryTimeoutMs: 4000
     },
     version: 1,
     createdAt: new Date(),
     updatedAt: new Date()
   })
   ```

## Rodando sem Docker
1. Garanta Mongo e Postgres rodando.
2. Exporte as variáveis de ambiente necessárias (veja `.env.example`).
3. Execute:

  ```sh
  go run ./cmd/api
  ```

## Endpoints
- `GET /health` — checagem simples.
- `POST /data/{source}/{table}` — executa SELECT com filtros e ordenação.

### Corpo da requisição
```json
{
  "schema": "public",        // opcional
  "limit": 100,               // default 100, limitado pelo datasource
  "offset": 0,                // default 0
  "countTotal": true,         // retorna total via COUNT(*)
  "orderBy": [
    { "field": "createdAt", "direction": "desc" }
  ],
  "filter": {
    "role": { "$eq": "PILOT" },
    "createdAt": { "$gte": "2025-01-01" }
  }
}
```

### Resposta de exemplo
```json
{
  "data": [
    { "id": "67fb3c8f-2f6b-4b2e-b5f6-3e9c9efc1c89", "email": "user@example.com" }
  ],
  "metadata": {
    "rows": 1,
    "table": "User",
    "tookMs": 12,
    "total": 42
  }
}
```

## Notas
- Identificadores de tabela/coluna são validados (letras, números, underscore) e escapados para Postgres.
- `limit` padrão é 100 e não passa de 500, ou do `maxRows` configurado no datasource.
- Valores de UUID e `time` retornam formatados como string.
