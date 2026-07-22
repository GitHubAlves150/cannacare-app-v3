# cannacare-app-v3

## 🗺️ ROADMAP DETALHADO

| Etapa | Módulo | Duração Estimada | Status |
| :---: | :--- | :---: | :---: |
| 1 | Configuração Inicial (Banco + GORM) | 1 dia | ✅ Concluído |
| 2 | Models + Migrations (Todas as tabelas) | 2 dias | ⏳ Próximo |
| 3 | Autenticação (JWT + Login/Register) | 2 dias | ⏳ |
| 4 | CRUD de Médicos | 1 dia | ⏳ |
| 5 | CRUD de Pacientes | 2 dias | ⏳ |
| 6 | Upload de Documentos | 2 dias | ⏳ |
| 7 | Gestão de Receitas/Prescrições | 2 dias | ⏳ |
| 8 | Sistema de Acolhimento (Anamnese) | 2 dias | ⏳ |
| 9 | CRUD de Produtos | 1 dia | ⏳ |
| 10 | Controle de Estoque (Lotes + Movimentações) | 2 dias | ⏳ |
| 11 | Sistema de Pedidos (com baixa de estoque) | 2 dias | ⏳ |
| 12 | Financeiro (Anuidades + Pagamentos) | 2 dias | ⏳ |
| 13 | Dashboard + Relatórios (Views) | 2 dias | ⏳ |
| 14 | Middleware (Roles + Permissões) | 1 dia | ⏳ |
| 15 | Testes + Documentação Final | 2 dias | ⏳ |


 ## 📝 O QUE CADA ETAPA VAI ENTREGAR:

Etapa 1: ✅ CONFIGURAÇÃO INICIAL (Concluída)

    Estrutura de pastas

    Configuração do GORM

    Conexão com PostgreSQL

    Health check

```bash
# Branch principal (produção)
main

# Branches por etapa
git checkout -b etapa-01-configuracao-inicial
git checkout -b etapa-02-models-migrations
git checkout -b etapa-03-autenticacao
git checkout -b etapa-04-medicos
git checkout -b etapa-05-pacientes
git checkout -b etapa-06-documentos
git checkout -b etapa-07-prescricoes
git checkout -b etapa-08-anamnese
git checkout -b etapa-09-produtos
git checkout -b etapa-10-estoque
git checkout -b etapa-11-pedidos
git checkout -b etapa-12-financeiro
git checkout -b etapa-13-dashboard
git checkout -b etapa-14-middleware
git checkout -b etapa-15-testes

```
## 📁 ETAPA 10: Controle de estoque(Lotes e movimentações)

Objetivo desta etapa:
    Nesta etapa é implementada o sistema de controle de estoque com lotes e movimentações. Esta é uma parte crucial para rastrear, controlar validações e gerar alertas.

📁 ESTRUTURA QUE VAMOS CRIAR


```bash
cannacare-app-v3/
├── internal/
│   ├── models/
│   │   ├── product_lot.go          # ✅ Já existe
│   │   └── stock_movement.go       # ✅ Já existe
│   ├── services/
│   │   ├── ... (todos os services existentes)
│   │   └── stock_service.go        # 🆕 Lógica para estoque
│   ├── handlers/
│   │   ├── ... (todos os handlers existentes)
│   │   └── stock_handler.go        # 🆕 Endpoints para estoque
│   ├── middleware/
│   │   └── auth.go                 # ✅ Já existe
│   └── utils/
│       ├── response.go             # ✅ Já existe
│       └── validators.go           # ✅ Já existe
├── pkg/
│   └── jwt/
│       └── jwt.go                  # ✅ Já existe
└── cmd/api/main.go                 # 🔄 Vamos atualizar


```

## ✅ TESTAR OS ENDPOINTS DE ESTOQUE

1. Fazer login

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cannacare.com","password":"admin123"}' \
  | jq -r '.data.token')

``` 

2. Criar um lote

```bash
PRODUCT_ID="id-do-produto-que-voce-criou"

curl -X POST http://localhost:8080/api/stock/lots \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "'$PRODUCT_ID'",
    "lot_number": "LOTE-001",
    "expiration_date": "2027-12-31T00:00:00Z",
    "quantity": 100,
    "supplier": "Fornecedor ABC",
    "purchase_date": "2026-07-01T00:00:00Z",
    "purchase_price": 120.50
  }' | jq '.'

``` 

3. Listar lotes
```bash
curl -X GET "http://localhost:8080/api/stock/lots?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
``` 

4. Ajustar estoque
```bash
LOT_ID="430009b4-b384-4bac-94d8-7dd0a28b672a"

curl -X POST http://localhost:8080/api/stock/adjust \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_lot_id": "'$LOT_ID'",
    "quantity": -10,
    "reason": "Ajuste por perda de produto"
  }' | jq '.'
``` 
5. Listar movimentações

```bash
curl -X GET "http://localhost:8080/api/stock/movements" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

``` 
6. Lotes com validade próxima
```bash
curl -X GET "http://localhost:8080/api/stock/expiring" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
``` 


xxxx

```bash
xxxxx

``` 

xxxx

```bash
xxxxx

``` 

xxxx

```bash
xxxxx

``` 
