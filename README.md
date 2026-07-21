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
## 📁 ETAPA 1: CONFIGURAÇÃO INICIAL

Objetivo desta etapa:

    Configurar a estrutura base do projeto, conexão com banco de dados usando GORM e um health check para validar que tudo está funcionando.

```bash

cannacare-backend/
├── cmd/
│   └── api/
│       └── main.go          # Ponto de entrada da aplicação (inicia o servidor)
├── internal/                 # Código interno (não pode ser importado por outros projetos)
│   ├── config/
│   │   └── config.go        # Carrega variáveis de ambiente (.env)
│   └── database/
│       └── db.go            # Gerencia a conexão com PostgreSQL usando GORM
├── pkg/                      # Código reutilizável (pode ser importado por outros projetos)
├── .env                      # Variáveis de ambiente (não versionar no Git!)
├── go.mod                    # Gerenciador de dependências
└── go.sum                    # Checksums das dependências (gerado automaticamente)

```

# Inicializa o módulo Go com o nome do projeto
# O nome será usado para importar pacotes internos
- go mod init cannacare-backend
- Para testar; go run cmd/api/main.go
- Acessar http://localhost:8080/
- http://localhost:8080/health

```bash
┌─────────────────────────────────────────────────────────────────┐
│                         SUA MÁQUINA                            │
│                                                                 │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │          CONTAINER DOCKER (cannacare_postgres)          │  │
│   │                                                         │  │
│   │   PostgreSQL 15 Alpine ✅                               │  │
│   │   Porta: 5433 (mapeada para 5432 interno)               │  │
│   │   Banco: cannacare_db ✅                                │  │
│   │   Usuário: postgres ✅                                  │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                │                                │
│                                ▼                                │
│                    🔗 Conexão estabelecida!                     │
│                                │                                │
│                                ▼                                │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │            APLICAÇÃO GO (RODANDO NA MÁQUINA)            │  │
│   │                                                         │  │
│   │   go run cmd/api/main.go ✅                             │  │
│   │   Porta da API: 8080 ✅                                 │  │
│   │   Health check: /health ✅                             │  │
│   └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```
