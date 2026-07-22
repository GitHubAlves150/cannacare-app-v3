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
## 📁 ETAPA 13: DashBoard de relatórios

Objetivo desta etapa:
    Nesta etapa será sistematizado o sistema de dashboard e relatórios. esta etapa é importante por que a gestão da associação precisa emitir relatórios frequentes, fornecendo visibilidade sobre todos os aspectos do negócio.

📁 ESTRUTURA QUE VAMOS CRIAR
```bash
cannacare-app-v3/
├── internal/
│   ├── services/
│   │   ├── ... (todos os services existentes)
│   │   └── dashboard_service.go    # 🆕 Lógica para dashboard
│   ├── handlers/
│   │   ├── ... (todos os handlers existentes)
│   │   └── dashboard_handler.go    # 🆕 Endpoints para dashboard
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

## ✅ TESTAR OS ENDPOINTS DE DASHBOARD

Fazer login

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cannacare.com","password":"admin123"}' \
  | jq -r '.data.token')

  # Ou

  curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cannacare.com","password":"admin123"}' | jq '.'
  
  

``` 
1. Visão geral do sistema
```bash

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMjkzNmRiM2UtY2FhMy00YjM0LTk5OWItMTViNDM0Y2EzMzE1IiwiZW1haWwiOiJhZG1pbkBjYW5uYWNhcmUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzg0ODQwNjI1LCJuYmYiOjE3ODQ3NTQyMjUsImlhdCI6MTc4NDc1NDIyNX0.x46WLxE1WFTipiQEFoEsmV-8XdFf-6lRGLT7kKDiPZw"

curl -X GET "http://localhost:8080/api/dashboard/overview" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
``` 
![alt text](image-11.png)

2. Relatório de pacientes
```bash

curl -X GET "http://localhost:8080/api/dashboard/patients" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
``` 
3. Receitas vencidas
```bash

curl -X GET "http://localhost:8080/api/dashboard/expired-prescriptions" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

``` 
4. Médicos que mais prescrevem
```bash
curl -X GET "http://localhost:8080/api/dashboard/top-doctors" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'

``` 
5. Produtos com estoque baixo
```bash

curl -X GET "http://localhost:8080/api/dashboard/low-stock" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
``` 
xxxx
```bash

xxxxxxx

``` 
xxxx
```bash

xxxxxxx

``` 
xxxx
```bash

xxxxxxx

``` 

