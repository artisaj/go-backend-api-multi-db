# API Database - Gerenciador de API Keys

## Overview
Sistema completo de gerenciamento de API Keys com interface web. As chaves são usadas para autenticar requisições à API de dados.

## Recursos

### 🔑 Gerenciamento de Chaves
- **Criar Chaves**: Gera automaticamente UUIDs v4 (não editáveis)
- **Associar Descrição**: Cada chave pode ter um nome (código da aplicação) e descrição
- **Copiar para Clipboard**: Botão rápido para copiar a chave
- **Editar**: Alterar nome e descrição da chave
- **Deletar**: Remover chaves com confirmação de segurança
- **Listar**: Ver todas as chaves criadas (com valores parcialmente obfuscados na listagem)

### 🌐 Interface Web
Acesse a página de gerenciamento em: `http://localhost:8080/keys.html`

## API Endpoints

### POST /api-keys
Criar uma nova chave API

**Request:**
```json
{
  "name": "Mobile App",
  "description": "App móvel iOS/Android",
  "permissions": []
}
```

**Response:**
```json
{
  "key": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Mobile App",
  "description": "App móvel iOS/Android",
  "permissions": [],
  "createdAt": "2025-12-31T22:24:22Z",
  "updatedAt": "2025-12-31T22:24:22Z"
}
```

### GET /api-keys
Listar todas as chaves (valores obfuscados)

**Response:**
```json
[
  {
    "key": "550e8400...",
    "name": "Mobile App",
    "description": "App móvel iOS/Android",
    "createdAt": "2025-12-31T22:24:22Z"
  }
]
```

### GET /api-keys/me
Obter informações da chave autenticada (requer header X-API-Key)

**Request:**
```
GET /api-keys/me
Headers: X-API-Key: 550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "key": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Mobile App",
  "description": "App móvel iOS/Android",
  "permissions": [],
  "createdAt": "2025-12-31T22:24:22Z",
  "updatedAt": "2025-12-31T22:24:22Z"
}
```

### PUT /api-keys/{key}
Atualizar nome e descrição de uma chave

**Request:**
```json
{
  "name": "Mobile App v2",
  "description": "App móvel iOS/Android - versão 2"
}
```

### DELETE /api-keys/{key}
Deletar uma chave API

**Response:** 204 No Content

## Usando as Chaves

### 1. Gerar Nova Chave
1. Acesse `http://localhost:8080/keys.html`
2. Clique no botão "+ Nova Chave"
3. Digite o nome da aplicação e descrição
4. Clique em "Salvar Chave"
5. A chave será gerada automaticamente (UUID v4)

### 2. Copiar Chave
1. Localize a chave desejada na lista
2. Clique no botão "Copiar"
3. A chave será copiada para o clipboard

### 3. Usar em Requisições
```bash
# Exemplo com curl
curl -X POST http://localhost:8080/data/racehub/Users \
  -H "X-API-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{"filter": {}}'
```

### 4. Atualizar Descrição
1. Localize a chave desejada
2. Clique no botão "Editar"
3. Altere nome/descrição
4. Clique em "Salvar"

### 5. Deletar Chave
1. Localize a chave desejada
2. Clique no botão "Deletar"
3. Confirme a exclusão
4. A chave será removida permanentemente

## Segurança

⚠️ **Importante:**
- As chaves são UUIDs v4 não editáveis - não as compartilhe
- Armazene em variáveis de ambiente, não no código
- Implemente rate limiting para evitar brute force (recomendado)
- Revise e delete chaves não utilizadas regularmente
- Use permissões granulares para limitar acesso

## Autenticação

Todas as requisições que requerem autenticação devem incluir o header:
```
X-API-Key: <sua-chave-api>
```

Se nenhuma chave for fornecida:
```json
{
  "code": "NO_API_KEY",
  "message": "no API key provided"
}
```

## Próximos Passos

- [ ] Implementar permissões granulares (database/table/column level)
- [ ] Dashboard de uso por chave
- [ ] Expiração automática de chaves
- [ ] Auditoria de requisições por chave
- [ ] Geração de novas chaves a partir das antigas (key rotation)
