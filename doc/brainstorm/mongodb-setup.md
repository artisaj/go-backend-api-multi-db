# Setup do MongoDB com DBeaver

## Docker Compose Setup

Executar:

```bash
docker-compose up -d
```

Isso inicia:

- **MongoDB** em `localhost:27017`
- **API Go** em `localhost:8080`

## Credenciais do MongoDB

```
Host: localhost
Port: 27017
Username: admin
Password: admin123
Database: api_database_config
```

## Conectar via MongoDB Compass

### Passo 1: Abrir o Compass e criar nova conexão

1. Clique em "New Connection".
2. Cole a connection string:
  ```
  mongodb://admin:admin123@localhost:27017/?authSource=admin
  ```
3. Clique em "Connect".

### Passo 2: Criar o database (se não aparecer)

- Clique em **Create database**.
- Database name: `api_database_config`
- Collection name: `data_sources`
- Clique em "Create". (Ele só aparece após existir pelo menos uma collection.)

### Passo 3: Inserir dados

- Abra o database `api_database_config`.
- Se precisar, crie a collection `data_sources` em "Create collection".
- Clique em `data_sources` → "Insert Document" → cole o JSON de exemplo e salve.

Coleções esperadas:
- `data_sources` (configuração de fontes)
- `query_metrics` (histórico de performance)
- `jobs` (consultas assíncronas)

## Coleção `data_sources`

Estrutura de exemplo para registrar um banco PostgreSQL:

```json
{
  "_id": ObjectId("..."),
  "name": "seguranca",
  "type": "postgres",
  "description": "Base de dados de segurança",
  "connection": {
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "postgres",
    "database": "seguranca"
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
  "createdAt": ISODate("2025-12-31T17:00:00Z"),
  "updatedAt": ISODate("2025-12-31T17:00:00Z")
}
```

### Inserir manualmente via DBeaver

1. Abrir a coleção `data_sources`
2. Clicar em "+" ou "Insert Document"
3. Colar JSON acima com seus dados reais
4. Salvar

## Health Check

```bash
curl http://localhost:8080/health
```

Resposta esperada:

```json
{ "status": "ok", "env": "development" }
```
