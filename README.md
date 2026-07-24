# cannacare-app-v3

## 🗺️ ROADMAP DETALHADO

| Etapa | Módulo | Duração Estimada | Status |
| :---: | :--- | :---: | :---: |
| 1 | Configuração Inicial (Banco + GORM) | 1 dia | ✅ Concluído |
| 2 | Models + Migrations (Todas as tabelas) | 2 dias | ✅ Concluído |
| 3 | Autenticação (JWT + Login/Register) | 2 dias | ✅ Concluído |
| 4 | CRUD de Médicos | 1 dia | ✅ Concluído |
| 5 | CRUD de Pacientes | 2 dias | ✅ Concluído |
| 6 | Upload de Documentos | 2 dias | ✅ Concluído |
| 7 | Gestão de Receitas/Prescrições | 2 dias | ✅ Concluído |
| 8 | Sistema de Acolhimento (Anamnese) | 2 dias | ✅ Concluído |
| 9 | CRUD de Produtos | 1 dia | ✅ Concluído |
| 10 | Controle de Estoque (Lotes + Movimentações) | 2 dias | ✅ Concluído |
| 11 | Sistema de Pedidos (com baixa de estoque) | 2 dias | ✅ Concluído |
| 12 | Financeiro (Anuidades + Pagamentos) | 2 dias | ✅ Concluído |
| 13 | Dashboard + Relatórios (Views) | 2 dias | ✅ Concluído |
| 14 | Middleware (Roles + Permissões) | 1 dia | ✅ Concluído |
| 15 | Testes + Documentação Final | 2 dias | ✅ Concluído |


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
## 📁 ETAPA 15: testes automatizados

Objetivo desta etapa:
    Reservei esta etapa para criar testes automatizados.

```bash
# Todas as rotas funcionando:
✅ /api/auth/*        - Autenticação
✅ /api/patients/*    - Pacientes
✅ /api/doctors/*     - Médicos
✅ /api/prescriptions/* - Receitas
✅ /api/orders/*      - Pedidos
✅ /api/stock/*       - Estoque
✅ /api/financial/*   - Financeiro
✅ /api/dashboard/*   - Dashboard
✅ /api/documents/*   - Documentos
✅ /api/products/*    - Produtos
✅ /api/anamnesis/*   - Acolhimento
``` 