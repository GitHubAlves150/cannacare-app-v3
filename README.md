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
## 📁 ETAPA 11: Sistema de pedidos

Objetivo desta etapa:
    O objetivo desta etapa é criar o sistema completo de pedidos, que é o coração do negócio. Está etapa integra pacientes, receitas, produtos e estoque.


📁 ESTRUTURA QUE VAMOS CRIAR

```bash

cannacare-app-v3/
├── internal/
│   ├── models/
│   │   ├── order.go              # ✅ Já existe
│   │   └── order_item.go         # ✅ Já existe
│   ├── services/
│   │   ├── ... (todos os services existentes)
│   │   └── order_service.go      # 🆕 Lógica para pedidos
│   ├── handlers/
│   │   ├── ... (todos os handlers existentes)
│   │   └── order_handler.go      # 🆕 Endpoints para pedidos
│   ├── middleware/
│   │   └── auth.go               # ✅ Já existe
│   └── utils/
│       ├── response.go           # ✅ Já existe
│       └── validators.go         # ✅ Já existe
├── pkg/
│   └── jwt/
│       └── jwt.go                # ✅ Já existe
└── cmd/api/main.go               # 🔄 Vamos atualizar

```

## ✅ TESTAR OS ENDPOINTS DE PEDIDOS

Fazer login

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cannacare.com","password":"admin123"}' \
  | jq -r '.data.token')

``` 

1. Criar um pedido
Primeiro, pegue os IDs necessários:
- Paciente aprovado 0f40331c-ca06-43bb-8618-06d64ee58e7b
- Prescrição válida 2517459c-8d34-41ea-9319-d0bcd9136f05
- Lote com estoque disponível 430009b4-b384-4bac-94d8-7dd0a28b672a  / 253166ea-38de-46f5-bb94-cdcea3f4d84b


```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMjkzNmRiM2UtY2FhMy00YjM0LTk5OWItMTViNDM0Y2EzMzE1IiwiZW1haWwiOiJhZG1pbkBjYW5uYWNhcmUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzg0ODM0NzE4LCJuYmYiOjE3ODQ3NDgzMTgsImlhdCI6MTc4NDc0ODMxOH0.Hudp65K1QasCh1oaZrLPJQmVl2cj1vJA4Y6b7x9bJEc"
PATIENT_ID="0f40331c-ca06-43bb-8618-06d64ee58e7b"
PRESCRIPTION_ID="2517459c-8d34-41ea-9319-d0bcd9136f05"
LOT_ID="430009b4-b384-4bac-94d8-7dd0a28b672a"

curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "patient_id": "'$PATIENT_ID'",
    "prescription_id": "'$PRESCRIPTION_ID'",
    "items": [
      {
        "product_lot_id": "'$LOT_ID'",
        "quantity": 2,
        "unit_price": 150.00
      }
    ],
    "notes": "Pedido de teste"
  }'

``` 

2. Atualizar status do pedido
```bash
TOKEN="seu-token"
ORDER_ID="id-do-seu-pedido"
curl -X PATCH "http://localhost:8080/api/orders/$ORDER_ID/status" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "correio",
    "notes": "Produto enviado pelo correio"
  }' | jq '.'

``` 
![alt text](image-6.png)


3. Gerar etiqueta
```bash
curl -X POST "http://localhost:8080/api/orders/$ORDER_ID/label" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

``` 
![alt text](image-7.png)

4. Adicionar rastreio
```bash

ORDER_ID="51a4542a-8ddf-48b6-9381-8c7b4d0ab8e6"

curl -X PATCH "http://localhost:8080/api/orders/$ORDER_ID/tracking" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tracking_code": "BR123456789BR",
    "shipping_carrier": "Correios"
  }' | jq '.'

``` 
![alt text](image-8.png)

## (Opcional) Atualizar para "entregue"
Quando o paciente receber o produto:
```bash
curl -X PATCH "http://localhost:8080/api/orders/$ORDER_ID/status" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "entregue",
    "notes": "Pedido entregue ao paciente"
  }' | jq '.'
```
![alt text](image-9.png)

5. Listar pedidos de um paciente
```bash
curl -X GET "http://localhost:8080/api/orders/patient/$PATIENT_ID" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

``` 

