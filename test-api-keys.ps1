#!/usr/bin/env pwsh

# Script de teste para API Keys Management
# Demonstra todo o fluxo CRUD de gerenciamento de chaves API

$baseUrl = "http://localhost:8080"
$apiKeysUrl = "$baseUrl/api-keys"

Write-Host "=== API Keys Management - Teste Completo ===" -ForegroundColor Cyan
Write-Host ""

# 1. Criar uma nova chave
Write-Host "1️⃣  Criando nova chave API..." -ForegroundColor Yellow
$createBody = @{
    name = "Test Application $(Get-Date -Format 'HHmmss')"
    description = "Aplicação de teste para validação do CRUD"
    permissions = @(
        @{
            resource = "users:*:*"
            level = "read"
        }
    )
} | ConvertTo-Json

$createResponse = Invoke-WebRequest -Uri $apiKeysUrl `
    -Method POST `
    -Headers @{'Content-Type'='application/json'} `
    -Body $createBody `
    -UseBasicParsing

$createdKey = $createResponse.Content | ConvertFrom-Json
Write-Host "✅ Chave criada com sucesso!" -ForegroundColor Green
Write-Host "   Key: $($createdKey.key)" -ForegroundColor Cyan
Write-Host "   Name: $($createdKey.name)" -ForegroundColor Cyan
Write-Host ""

# 2. Listar todas as chaves
Write-Host "2️⃣  Listando todas as chaves..." -ForegroundColor Yellow
$listResponse = Invoke-WebRequest -Uri $apiKeysUrl -UseBasicParsing
$keysList = $listResponse.Content | ConvertFrom-Json
Write-Host "✅ Total de chaves: $($keysList.Count)" -ForegroundColor Green
$keysList | ForEach-Object {
    Write-Host "   - $($_.name) (criada em: $($_.createdAt))" -ForegroundColor Cyan
}
Write-Host ""

# 3. Obter a chave autenticada
Write-Host "3️⃣  Testando autenticação com /api-keys/me..." -ForegroundColor Yellow
try {
    $meResponse = Invoke-WebRequest -Uri "$apiKeysUrl/me" `
        -Headers @{'X-API-Key'=$createdKey.key} `
        -UseBasicParsing
    $meKey = $meResponse.Content | ConvertFrom-Json
    Write-Host "✅ Autenticação funcionou!" -ForegroundColor Green
    Write-Host "   Name: $($meKey.name)" -ForegroundColor Cyan
    Write-Host "   Permissions: $($meKey.permissions.Count)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Erro na autenticação" -ForegroundColor Red
}
Write-Host ""

# 4. Atualizar a chave
Write-Host "4️⃣  Atualizando descrição da chave..." -ForegroundColor Yellow
$updateBody = @{
    name = $createdKey.name
    description = "Atualizada em $(Get-Date -Format 'dd/MM/yyyy HH:mm:ss')"
} | ConvertTo-Json

$updateResponse = Invoke-WebRequest -Uri "$apiKeysUrl/$($createdKey.key)" `
    -Method PUT `
    -Headers @{'Content-Type'='application/json'} `
    -Body $updateBody `
    -UseBasicParsing

$updatedKey = $updateResponse.Content | ConvertFrom-Json
Write-Host "✅ Chave atualizada!" -ForegroundColor Green
Write-Host "   Description: $($updatedKey.description)" -ForegroundColor Cyan
Write-Host ""

# 5. Testar acesso sem API Key
Write-Host "5️⃣  Testando acesso sem API Key (deve falhar)..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri "$apiKeysUrl/me" -UseBasicParsing
    Write-Host "❌ Erro: acesso foi permitido sem chave!" -ForegroundColor Red
} catch {
    Write-Host "✅ Acesso corretamente negado!" -ForegroundColor Green
    $error = $_.Exception.Response.StatusCode
    Write-Host "   Status: $error" -ForegroundColor Cyan
}
Write-Host ""

# 6. Deletar a chave de teste
Write-Host "6️⃣  Deletando a chave de teste..." -ForegroundColor Yellow
$deleteResponse = Invoke-WebRequest -Uri "$apiKeysUrl/$($createdKey.key)" `
    -Method DELETE `
    -UseBasicParsing

Write-Host "✅ Chave deletada com sucesso!" -ForegroundColor Green
Write-Host ""

# 7. Verificar deleção
Write-Host "7️⃣  Verificando deleção..." -ForegroundColor Yellow
$finalList = Invoke-WebRequest -Uri $apiKeysUrl -UseBasicParsing
$finalKeysList = $finalList.Content | ConvertFrom-Json
$keyExists = $finalKeysList | Where-Object { $_.key -eq $createdKey.key }
if ($keyExists) {
    Write-Host "❌ Erro: chave não foi deletada!" -ForegroundColor Red
} else {
    Write-Host "✅ Chave foi removida com sucesso!" -ForegroundColor Green
}
Write-Host ""

Write-Host "=== Teste Concluído com Sucesso! ===" -ForegroundColor Green
