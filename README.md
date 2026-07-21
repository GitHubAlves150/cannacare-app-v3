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
## 📁 ETAPA 2: Models + migratioons

Objetivo desta etapa:

    Mapeamento das tabelas em GO e rodar Migrations para criar as tabelas no PostgreSQO

```bash

cannacare-backend/
├── internal/
│   ├── models/
│   │   ├── user.go              # Usuários do sistema
│   │   ├── patient.go           # Pacientes
│   │   ├── doctor.go            # Médicos prescritores
│   │   ├── prescription.go      # Receitas médicas
│   │   ├── prescription_item.go # Itens da receita
│   │   ├── product.go           # Produtos (óleos)
│   │   ├── product_lot.go       # Lotes de produtos
│   │   ├── stock_movement.go    # Movimentações de estoque
│   │   ├── order.go             # Pedidos
│   │   ├── order_item.go        # Itens do pedido
│   │   ├── payment.go           # Pagamentos
│   │   ├── subscription.go      # Anuidades
│   │   ├── anamnese.go          # Acolhimento/rastreamento
│   │   ├── patient_document.go  # Documentos do paciente
│   │   ├── notification.go      # Notificações
│   │   └── patient_status_history.go # Histórico de status
│   └── database/
│       └── migrate.go           # Script de migração

```
