# Brainstorm inicial

## Objetivo

Centralizar acesso a dados de múltiplos bancos (PostgreSQL, MongoDB, etc.) via API unificada escrita em Go, com caching inteligente (query-hash + entity-version) e fallback para consultas assíncronas (query-id) quando o custo/tempo exceder limites.

## Arquitetura (Clean/Hexagonal)

- **Presentation**: rotas HTTP (chi), middlewares (request-id, logging, timeout).
- **Application**: casos de uso, orquestração, DTO/validação, cálculo de custo de consulta.
- **Domain**: entidades, serviços de domínio, portas (repos, cache, fila, telemetria).
- **Infrastructure**: adaptadores (connectors DB, cache, fila, telemetry), configs; client Mongo para metadados.
- **Config**: `MongoDB` guarda metadata de conexões; env overrides para segredos (sem vault por enquanto, TODO futuro).

## Configuração de fontes de dados (Mongo collection)

- Campos: `id`, `name`, `type` (postgres|mongodb|dynamodb|...); `connection` (host, port, user, password, db); `capabilities` (joins, projections, depthLimit); `limits` (maxRows, timeoutMs); `version` (para cache busting); `createdAt/updatedAt`.
- Dev: credenciais em texto (env); Prod: TODO usar secrets manager.

## Conectores (Strategy + Factory)

- Porta `DatabaseConnector` (Go interface): `Query`, `Execute`, `Transaction`, `Close`.
- Implementações: Postgres (pgx), Mongo (mongo driver), outros via fábrica.
- Query Builder/Translator por tipo (SQL builder com binds; Mongo filter builder).

## Caching inteligente

- Chave = `query-hash + entity-version` (version incrementa em mutações/ingestões) para cache busting automático.
- TTL + invalidação por eventos (Domain Events -> Event Bus -> Cache handler).
- Política de tamanho: limite de itens e LRU/score; cache é caro — guardar apenas payloads pequenos/frequentes para melhorar fluxo, não para arquivar dados.
- Consultas imutáveis: opção de TTL longo/manual, mas com quota de cache por tabela/entidade.

## Estimativa de custo de consulta

- Persistir métricas históricas por `{fonte}/{tabela}/{shape-de-query}`: p50/p95/timeout count, rows retornadas, tamanho médio.
- Antes de executar: estimar custo; se previsão > `queryTimeoutMs` ou acima de threshold p95 -> enfileirar (async) e retornar `queryId`.
- Cache ajuda na predição: tempos anteriores (hit/miss) entram no modelo simples.

## Consultas longas (query-id)

- Endpoint principal: `POST /data/{fonte}/{tabela}` com filtros/order/include/depth-limit.
- Se custo alto: criar job `pending`, retornar `queryId`; worker processa, salva resultado resumido; endpoint `GET /queries/{queryId}` retorna status (`pending|processing|completed|failed`) e resultado quando pronto.
- Limpeza por TTL de jobs e resultados; opcional arquivar histórico.

## Observabilidade day-one

- Correlation/Request ID middleware; logger estruturado (zerolog) com request-id.
- Tracing: spans em conectores, cache, fila, use-cases.
- Métricas: latência de query (p50/p95), cache hit/miss, duração de jobs, erros por conector, profundidade de include.

## Segurança (dev simplificado)

- CORS/Helmet básico. Dev: credenciais abertas em env. TODO: secrets manager e auth de chamadas.

## Próximos passos (alto nível)

1. Modelar contratos das portas (cache, queue, database, telemetry, config repository).
2. Implementar config loader (Mongo) e factory de conectores.
3. Definir DTO/validação para `POST /data/{fonte}/{tabela}` (filtros, order, include com depth-limit).
4. Implementar calculadora de custo e roteamento sync vs async.
5. Instrumentar logging/tracing/metrics mínimos.
