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
## 📁 ETAPA 7: gestãpo de prescriçôes(Receita médica).

Objetivo desta etapa:

   Nesta etapa será implementada o sistema de gestão de prescrições/receitas médicas.

📁 ESTRUTURA QUE VAMOS CRIAR.

```bash

cannacare-app-v3/
├── internal/
│   ├── models/
│   │   ├── prescription.go           # ✅ Já existe
│   │   └── prescription_item.go      # ✅ Já existe
│   ├── services/
│   │   ├── auth_service.go           # ✅ Já existe
│   │   ├── doctor_service.go         # ✅ Já existe
│   │   ├── patient_service.go        # ✅ Já existe
│   │   ├── document_service.go       # ✅ Já existe
│   │   └── prescription_service.go   # 🆕 Lógica para prescrições
│   ├── handlers/
│   │   ├── auth_handler.go           # ✅ Já existe
│   │   ├── doctor_handler.go         # ✅ Já existe
│   │   ├── patient_handler.go        # ✅ Já existe
│   │   ├── document_handler.go       # ✅ Já existe
│   │   └── prescription_handler.go   # 🆕 Endpoints para prescrições
│   ├── middleware/
│   │   └── auth.go                   # ✅ Já existe
│   └── utils/
│       ├── response.go               # ✅ Já existe
│       └── validators.go             # ✅ Já existe
├── pkg/
│   └── jwt/
│       └── jwt.go                    # ✅ Já existe
└── cmd/api/main.go                   # 🔄 Vamos atualizar

```

## ✅ TESTAR OS ENDPOINTS DE PRESCRIÇÕES.

1. Como não foi criado um produto ainda(Etapa 9), vamos criar um diretamento no banco.
```bash
# Conectar ao banco
docker exec -it cannacare_postgres psql -U postgres -d cannacare_db

# Inserir um produto
INSERT INTO products (id, name, description, unit_price, min_stock_alert, is_active, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'Óleo CBD Full Spectrum 10% - 30ml',
    'Óleo com 10% de CBD, 30ml, uso sublingual',
    150.00,
    10,
    true,
    NOW(),
    NOW()
);

``` 
## 📝 PASSO 2: APROVAR UM PACIENTE.
```bash

# 2.1. Listar pacientes para ver os IDs
curl -X GET "http://localhost:8080/api/patients?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```
```bash
# 2.2. Aprovar o paciente (mudar status para "aprovado")
PATIENT_ID="id-do-paciente-que-voce-criou"

curl -X PATCH "http://localhost:8080/api/patients/$PATIENT_ID/status" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "aprovado",
    "reason": "Paciente aprovado para teste"
  }
``` 
```bash
# 2.3. Verificar se o paciente foi aprovado
curl -X GET "http://localhost:8080/api/patients/$PATIENT_ID" \
  -H "Authorization: Bearer $TOKEN"

```

3. Criar uma prescrição
``` bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMjkzNmRiM2UtY2FhMy00YjM0LTk5OWItMTViNDM0Y2EzMzE1IiwiZW1haWwiOiJhZG1pbkBjYW5uYWNhcmUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzg0ODE5NDMyLCJuYmYiOjE3ODQ3MzMwMzIsImlhdCI6MTc4NDczMzAzMn0.A0ji8a55DnaIEmkd9vXYcf7wOHNfdq1-HptXRacW2Yo"

curl -X POST http://localhost:8080/api/prescriptions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "patient_id": "0f40331c-ca06-43bb-8618-06d64ee58e7b",
    "doctor_id": "fe063e1a-e653-437d-81cd-223d14c1eafe",
    "cid": "G40.0",
    "issue_date": "2026-01-01T00:00:00Z",
    "expiration_date": "2027-01-01T00:00:00Z",
    "items": [
      {
        "product_id": "ae13d4b8-35e9-4d0f-9aa9-d4679c5b4c56",
        "dosage_instructions": "5 gotas a cada 12 horas, sublingual",
        "quantity_recommended": 2
      }
    ]
  }'
```

## 2. Listar prescrições.

``` bash
curl -X GET "http://localhost:8080/api/prescriptions?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```


## 3. Validar prescrição.

``` bash
PRESCRIPTION_ID="2517459c-8d34-41ea-9319-d0bcd9136f05"

curl -X GET "http://localhost:8080/api/prescriptions/validate/$PRESCRIPTION_ID" \
  -H "Authorization: Bearer $TOKEN"
```

## 4. Buscar prescrições vencidas.

``` bash
curl -X GET "http://localhost:8080/api/prescriptions/expired" \
  -H "Authorization: Bearer $TOKEN"
```


## 5. Atualizar status em lote.

``` bash
curl -X POST "http://localhost:8080/api/prescriptions/update-status" \
  -H "Authorization: Bearer $TOKEN"
```
