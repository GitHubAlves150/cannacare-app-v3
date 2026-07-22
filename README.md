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
## 📁 ETAPA 14: Middleware e permissões(Finalização)

Objetivo desta etapa:
    Nesta etapa terminamos a parte de middlewares e permissões, garantindo que cada perfil de usuário tenha acesso apenas ao que realmente precisa. também vamos adicionar validações extras e logs.

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
xxxx
```bash
xxxxxx

``` 