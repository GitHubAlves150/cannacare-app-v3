📚 DOCUMENTAÇÃO COMPLETA DO CANNACARE

Aqui está a documentação completa para você salvar na branch do projeto. Esta documentação cobre todos os aspectos do sistema para que qualquer desenvolvedor possa entender e dar continuidade ao projeto.
📁 ESTRUTURA DA DOCUMENTAÇÃO
text
```bash
📁 cannacare-docs/
├── 📄 README.md                    # Visão geral do projeto
├── 📄 01-arquitetura.md             # Arquitetura do sistema
├── 📄 02-banco-de-dados.md          # Modelagem do banco
├── 📄 03-backend.md                 # Backend (Go + Chi + GORM)
├── 📄 04-frontend.md                # Frontend (Next.js + TypeScript)
├── 📄 05-multi-tenancy.md           # Multi-tenancy (associação)
├── 📄 06-autenticacao.md            # Autenticação e JWT
├── 📄 07-rotas-api.md               # Lista completa de endpoints
├── 📄 08-deploy.md                  # Guia de deploy
└── 📄 09-faq.md                     # Perguntas frequentes

📄 00-README.md
markdown

# 🌿 CannaCare - Sistema de Gestão para Associações de Cannabis Medicinal

## 📋 Visão Geral

O CannaCare é um **SaaS (Software as a Service)** multi-tenant para gestão de associações de cannabis medicinal. O sistema permite que associações gerenciem pacientes, médicos, receitas, estoque, pedidos e financeiro de forma integrada e segura.

## 🎯 Objetivo

Facilitar a gestão de associações de cannabis medicinal, oferecendo uma plataforma completa para:
- Acolhimento e gestão de pacientes
- Controle de receitas médicas
- Gestão de estoque de óleos e produtos
- Dispensação e pedidos
- Controle financeiro (anuidades e pagamentos)
- Relatórios e dashboards

## 🏗️ Tecnologias

| Camada | Tecnologia | Versão |
|--------|------------|--------|
| **Backend** | Go (Golang) | 1.21+ |
| **Framework HTTP** | Chi | v5 |
| **ORM** | GORM | v1.25 |
| **Banco de Dados** | PostgreSQL | 15+ |
| **Frontend** | Next.js | 14.x |
| **Linguagem Frontend** | TypeScript | 5.x |
| **Estilização** | Tailwind CSS | 3.x |
| **Autenticação** | JWT (golang-jwt) | v5 |
| **Containerização** | Docker | 24+ |

## 📊 Status do Projeto

| Módulo | Status |
|--------|--------|
| ✅ Banco de Dados | Concluído |
| ✅ Backend (API) | Concluído |
| ✅ Frontend | Concluído |
| ✅ Multi-Tenancy | Concluído |
| ✅ Autenticação | Concluído |
| ✅ Deploy | Pendente |

## 🚀 Como Rodar

### Pré-requisitos

- Docker e Docker Compose
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+ (ou via Docker)

### 1. Clonar o repositório

```bash
git clone git@github.com:GitHubAlves150/cannacare-app-v3.git
git clone git@github.com:GitHubAlves150/cannacare-app-V3-Front.git

2. Rodar o Backend
bash

cd cannacare-app-v3

# Criar arquivo .env
cat > .env << 'EOF'
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=cannacare2026!
DB_NAME=cannacare_db
DB_SSLMODE=disable
JWT_SECRET=cannacare-super-secret-key-2026
JWT_EXPIRES_IN=24h
SERVER_PORT=8080
ENV=development
EOF

# Subir o banco de dados (Docker)
docker compose up -d

# Rodar o backend
go run cmd/api/main.go

3. Rodar o Frontend
bash

cd cannacare-app-V3-Front/cannacare-frontend

# Instalar dependências
npm install

# Rodar o frontend
npm run dev

4. Acessar o sistema

    Frontend: http://localhost:3000

    Backend API: http://localhost:8080

5. Credenciais de acesso
Campo	Valor
Email	admin@associacao.com
Senha	123456
📁 Estrutura do Projeto
Backend
text

cannacare-app-v3/
├── cmd/
│   └── api/
│       └── main.go              # Ponto de entrada
├── internal/
│   ├── config/                   # Configurações
│   ├── database/                 # Conexão e migrações
│   ├── handlers/                 # Handlers HTTP (controllers)
│   ├── middleware/               # Middlewares (auth, roles, logger)
│   ├── models/                   # Models GORM
│   ├── services/                 # Lógica de negócio
│   └── utils/                    # Funções auxiliares
├── pkg/
│   └── jwt/                      # JWT Service
└── uploads/
    └── documents/                # Arquivos enviados

Frontend
text

cannacare-app-V3-Front/
├── app/
│   ├── dashboard/                # Páginas do dashboard
│   ├── login/                    # Página de login
│   └── register/                 # Página de registro
├── components/                   # Componentes reutilizáveis
├── lib/
│   └── api/                      # Chamadas para a API
└── public/                       # Arquivos estáticos

🔐 Multi-Tenancy

O sistema é multi-tenant, ou seja, suporta múltiplas associações (clientes) na mesma instância.
Como funciona:

    Cada associação tem um association_id único

    Todas as tabelas têm a coluna association_id

    O JWT contém o association_id do usuário

    O middleware extrai o association_id de cada requisição

    Todas as queries SQL filtram por association_id

Diagrama:
text

┌─────────────────────────────────────────────────────────────┐
│                     CANNACARE (SaaS)                       │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │
│  │ Associação A│  │ Associação B│  │ Associação C│       │
│  │ ID: 8e0b... │  │ ID: 9f1c... │  │ ID: a2d3... │       │
│  ├─────────────┤  ├─────────────┤  ├─────────────┤       │
│  │ Pacientes   │  │ Pacientes   │  │ Pacientes   │       │
│  │ Médicos     │  │ Médicos     │  │ Médicos     │       │
│  │ Receitas    │  │ Receitas    │  │ Receitas    │       │
│  │ Pedidos     │  │ Pedidos     │  │ Pedidos     │       │
│  └─────────────┘  └─────────────┘  └─────────────┘       │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  BANCO DE DADOS (Todas as tabelas têm association_id)│   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘

📋 Funcionalidades
👤 Pacientes

    Cadastro com documentos (RG/CPF, comprovante residência, laudo médico)

    Fila Regulatória (aprovação de pacientes)

    Status: pendente_documentacao → em_analise → aprovado/negado

    Paciente social (isenção de anuidade)

👨‍⚕️ Médicos

    Cadastro com CRM e especialidade

    Ativo/Inativo

💊 Receitas

    Múltiplos itens por receita

    Validação de receita

    Status: valida → proxima_vencer → vencida

    Upload de arquivo (receita scaneada)

🏥 Acolhimento (Anamnese)

    Anamnese inicial

    Rastreios periódicos (1 mês, 3 meses, 6 meses)

    Acompanhamento contínuo

📦 Estoque

    Produtos com preço e estoque mínimo

    Lotes com validade e quantidade

    Movimentações (entrada, baixa, ajuste, perda)

    Alertas de estoque baixo e validade próxima

🛒 Pedidos

    Validação de receita antes do pedido

    Status: pendente → separado → dispensa → correio → entregue

    Código de rastreio

    Baixa automática no estoque

💰 Financeiro

    Anuidades (anuais)

    Pagamentos (PIX, boleto, cartão, transferência)

    Tipos: anuidade, compra_produto, doacao

📊 Relatórios

    Visão geral do sistema

    Receitas vencidas

    Médicos que mais prescrevem

    Produtos com estoque baixo

    Anuidades em atraso

🔐 Perfis e Permissões
Role	Permissões
admin	Acesso total ao sistema
coordenacao	Relatórios, aprovações
secretaria	Pacientes, documentos, pagamentos
acolhimento	Anamnese, acompanhamento
farmacia	Estoque, pedidos
🐳 Docker
bash

# Subir banco de dados
docker compose up -d

# Ver containers
docker ps

# Acessar banco
docker exec -it cannacare_postgres psql -U postgres -d cannacare_db

📝 Licença

MIT License - veja o arquivo LICENSE
👤 Autor

Lucas Lorenço Alves

🌿 CannaCare - Gestão Inteligente para Associações de Cannabis Medicinal
text


---

## 📄 01-arquitetura.md

```markdown
# 🏗️ Arquitetura do Sistema

## Visão Geral

O CannaCare segue uma arquitetura **monolítica modular** com separação clara de responsabilidades:

┌─────────────────────────────────────────────────────────────────────────────┐
│ CLIENTE │
│ (Navegador / App Mobile) │
└─────────────────────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ FRONTEND (Next.js) │
│ • App Router │
│ • Server Components + Client Components │
│ • Tailwind CSS │
│ • Axios para chamadas API │
└─────────────────────────────────────────────────────────────────────────────┘
│
▼ (API REST)
┌─────────────────────────────────────────────────────────────────────────────┐
│ BACKEND (Go + Chi) │
│ ┌─────────────────────────────────────────────────────────────────────┐ │
│ │ CAMADA DE HANDLERS (HTTP) │ │
│ │ • Recebe requisições HTTP │ │
│ │ • Valida dados de entrada │ │
│ │ • Extrai association_id do Context │ │
│ │ • Retorna respostas padronizadas │ │
│ └─────────────────────────────────────────────────────────────────────┘ │
│ │ │
│ ┌─────────────────────────────────────────────────────────────────────┐ │
│ │ CAMADA DE SERVICES (Lógica de Negócio) │ │
│ │ • Regras de negócio │ │
│ │ • Validações │ │
│ │ • Transações │ │
│ │ • Filtro por association_id │ │
│ └─────────────────────────────────────────────────────────────────────┘ │
│ │ │
│ ┌─────────────────────────────────────────────────────────────────────┐ │
│ │ CAMADA DE MODELS (GORM) │ │
│ │ • Representação das tabelas │ │
│ │ • Relacionamentos │ │
│ │ • Hooks (BeforeCreate, BeforeUpdate) │ │
│ └─────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ BANCO DE DADOS (PostgreSQL) │
│ • 17 tabelas com association_id │
│ • 6 views para relatórios │
│ • Triggers e funções │
└─────────────────────────────────────────────────────────────────────────────┘
text


## 📋 Camadas Detalhadas

### 1. Camada de Handlers (HTTP)

**Responsabilidade:** Receber requisições HTTP, validar dados e retornar respostas.

```go
// internal/handlers/patient_handler.go
func (h *PatientHandler) List(w http.ResponseWriter, r *http.Request) {
    // 1. Extrair association_id do Context
    associationID := r.Context().Value(middleware.AssociationIDKey).(string)
    
    // 2. Processar requisição
    patients, total, err := h.patientService.List(associationID, req)
    
    // 3. Retornar resposta
    utils.SendSuccess(w, http.StatusOK, patients)
}

2. Camada de Services (Lógica de Negócio)

Responsabilidade: Implementar regras de negócio, validações e transações.
go

// internal/services/patient_service.go
func (s *PatientService) List(associationID uuid.UUID, req ListPatientRequest) ([]PatientResponse, int64, error) {
    // 1. SEMPRE filtrar por association_id
    query := s.db.Model(&models.Patient{}).Where("association_id = ?", associationID)
    
    // 2. Aplicar filtros
    if req.Name != "" {
        query = query.Where("full_name ILIKE ?", "%"+req.Name+"%")
    }
    
    // 3. Retornar resultados
    var patients []models.Patient
    query.Find(&patients)
    return patients, total, nil
}

3. Camada de Models (Banco de Dados)

Responsabilidade: Representar as tabelas do banco e seus relacionamentos.
go

// internal/models/patient.go
type Patient struct {
    BaseModel
    AssociationID uuid.UUID `gorm:"type:uuid;not null;index" json:"association_id"`
    FullName      string    `gorm:"not null" json:"full_name"`
    CPF           string    `gorm:"not null" json:"cpf"`
    // ...
}

4. Middleware

Responsabilidade: Processar requisições antes de chegarem aos handlers.
go

// internal/middleware/auth.go
func AuthMiddleware(jwtService *jwt.JWTService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Validar token JWT
            claims, err := jwtService.ValidateToken(token)
            
            // 2. Extrair association_id
            ctx := context.WithValue(r.Context(), AssociationIDKey, claims.AssociationID)
            
            // 3. Chamar próximo handler
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

🔐 Fluxo de Autenticação
text

1. Usuário envia email + senha para /api/auth/login
2. Backend valida credenciais
3. Backend gera JWT com association_id
4. Frontend armazena token no localStorage
5. Frontend envia token em todas as requisições (Header: Authorization: Bearer <token>)
6. Middleware valida token e extrai association_id
7. Todas as queries filtram por association_id

🗄️ Fluxo de Dados
text

1. Usuário faz requisição GET /api/patients
2. Middleware extrai association_id do JWT
3. Handler chama Service com association_id
4. Service filtra query por association_id
5. PostgreSQL retorna apenas pacientes da associação
6. Handler retorna resposta para o frontend

📊 Diagrama de Sequência (Criar Paciente)
text

Usuário → Frontend → Backend → Banco
   │          │          │          │
   │   POST /api/patients           │
   │──────────►│          │          │
   │          │          │          │
   │          │   POST /api/patients│
   │          │──────────►│          │
   │          │          │          │
   │          │          │   INSERT │
   │          │          │──────────►│
   │          │          │          │
   │          │          │   RETURN │
   │          │          │◄──────────│
   │          │          │          │
   │          │   201 Created       │
   │          │◄──────────│          │
   │          │          │          │
   │   201 Created       │          │
   │◄──────────│          │          │
   │          │          │          │

🛠️ Tecnologias Justificadas
Tecnologia	Motivo
Go	Performance, concorrência, fácil deploy
Chi	Leve, rápido, bom para APIs REST
GORM	Produtividade, migrations, hooks
PostgreSQL	Confiável, suporte a JSON, UUID
Next.js	SSR, SEO, App Router moderno
TypeScript	Segurança de tipos no frontend
Tailwind CSS	Estilização rápida e consistente
JWT	Stateless, escalável
🔒 Segurança
Camada	Medida
Autenticação	JWT com expiração (24h)
Senhas	Bcrypt (hash + salt)
SQL	GORM protege contra SQL Injection
CORS	Configurado para segurança
Dados	Isolamento por association_id
Soft Delete	DeletedAt para preservar histórico
📈 Escalabilidade
Aspecto	Estratégia
Banco de Dados	Índices em association_id
API	Stateless (fácil escalar horizontalmente)
Cache	Futuro: Redis para cache
CDN	Frontend via Vercel (CDN global)
text


---

## 📄 02-banco-de-dados.md

```markdown
# 🗄️ Banco de Dados - Modelagem

## 📊 Visão Geral

O banco de dados é **PostgreSQL 15+** com suporte a **UUID** e **JSONB**. O modelo é **multi-tenant**, com todas as tabelas tendo a coluna `association_id`.

## 🏗️ Diagrama de Relacionamentos

┌─────────────────────────────────────────────────────────────────────────────┐
│ ASSOCIATIONS │
│ id (PK) | name | cnpj | email | plan | status | patient_limit │
└─────────────────────────────────────────────────────────────────────────────┘
│ 1
│
│ N
▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ USERS │
│ id (PK) | association_id (FK) | name | email | password_hash | role │
└─────────────────────────────────────────────────────────────────────────────┘
│
│
▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ PATIENTS │
│ id (PK) | association_id (FK) | full_name | cpf | status │
└─────────────────────────────────────────────────────────────────────────────┘
│ 1 │ 1 │ 1
│ N │ N │ N
▼ ▼ ▼
┌───────────────┐ ┌─────────────────────┐ ┌─────────────────────────┐
│PATIENT_DOCS │ │ PRESCRIPTIONS │ │ ORDERS │
│patient_id(FK) │ │ patient_id(FK) │ │ patient_id(FK) │
│association_id │ │ doctor_id(FK) │ │ prescription_id(FK) │
└───────────────┘ │ association_id │ │ association_id │
└─────────────────────┘ └─────────────────────────┘
│ 1 │ 1
│ N │ N
▼ ▼
┌─────────────────┐ ┌─────────────────────────┐
│PRESCRIPTION_ITEMS│ │ ORDER_ITEMS │
│prescription_id(FK)│ │ order_id(FK) │
│product_id(FK) │ │ product_lot_id(FK) │
│association_id │ │ association_id │
└─────────────────┘ └─────────────────────────┘
text


## 📋 Tabelas

### 1. associations (MESTRA - Multi-Tenant)

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| name | VARCHAR(200) | Nome da associação |
| cnpj | VARCHAR(18) | CNPJ (único) |
| email | VARCHAR(200) | Email (único) |
| phone | VARCHAR(20) | Telefone |
| address | TEXT | Endereço |
| plan | VARCHAR(50) | basic, premium, enterprise |
| status | VARCHAR(20) | pending, active, suspended, cancelled |
| patient_limit | INTEGER | Limite de pacientes |
| trial_ends_at | TIMESTAMP | Fim do período de teste |
| stripe_customer_id | VARCHAR(100) | ID no Stripe |
| subscription_id | VARCHAR(100) | ID da assinatura |
| created_at | TIMESTAMP | Data de criação |
| updated_at | TIMESTAMP | Data de atualização |
| deleted_at | TIMESTAMP | Soft delete |

### 2. users

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| association_id | UUID | FK para associations |
| name | VARCHAR(200) | Nome do usuário |
| email | VARCHAR(200) | Email (único por associação) |
| password_hash | VARCHAR(255) | Hash da senha (bcrypt) |
| role | VARCHAR(50) | admin, secretaria, acolhimento, farmacia, coordenacao |
| is_active | BOOLEAN | Usuário ativo? |
| last_login_at | TIMESTAMP | Último login |
| created_at | TIMESTAMP | Data de criação |
| updated_at | TIMESTAMP | Data de atualização |
| deleted_at | TIMESTAMP | Soft delete |

**Constraint:** UNIQUE(email, association_id)

### 3. patients

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| association_id | UUID | FK para associations |
| user_id | UUID | FK para users (opcional) |
| full_name | VARCHAR(200) | Nome completo |
| birth_date | DATE | Data de nascimento |
| gender | VARCHAR(20) | Gênero |
| cpf | VARCHAR(14) | CPF (único por associação) |
| rg | VARCHAR(20) | RG |
| phone | VARCHAR(20) | Telefone |
| whatsapp | VARCHAR(20) | WhatsApp |
| email | VARCHAR(200) | Email |
| address_* | VARCHAR | Endereço completo |
| status | VARCHAR(50) | Status do paciente |
| is_social_patient | BOOLEAN | Paciente social? |
| social_assistant_notes | TEXT | Observações |
| approved_at | TIMESTAMP | Data de aprovação |
| created_at | TIMESTAMP | Data de criação |
| updated_at | TIMESTAMP | Data de atualização |
| deleted_at | TIMESTAMP | Soft delete |

**Status:** pendente_documentacao, em_analise, aprovado, negado, assistente_social

### 4. doctors

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| association_id | UUID | FK para associations |
| name | VARCHAR(200) | Nome do médico |
| crm | VARCHAR(20) | CRM (único por associação) |
| crm_state | VARCHAR(2) | UF do CRM |
| specialty | VARCHAR(100) | Especialidade |
| phone | VARCHAR(20) | Telefone |
| email | VARCHAR(200) | Email |
| is_active | BOOLEAN | Ativo? |
| created_at | TIMESTAMP | Data de criação |
| updated_at | TIMESTAMP | Data de atualização |
| deleted_at | TIMESTAMP | Soft delete |

### 5. prescriptions

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| association_id | UUID | FK para associations |
| patient_id | UUID | FK para patients |
| doctor_id | UUID | FK para doctors |
| cid | VARCHAR(10) | Código CID |
| issue_date | DATE | Data de emissão |
| expiration_date | DATE | Data de validade |
| status | VARCHAR(20) | valida, proxima_vencer, vencida |
| is_active | BOOLEAN | Ativa? |
| prescription_file_url | TEXT | URL do arquivo |
| prescription_file_name | VARCHAR(255) | Nome do arquivo |
| validated_by | UUID | FK para users |
| validated_at | TIMESTAMP | Data de validação |
| created_at | TIMESTAMP | Data de criação |
| updated_at | TIMESTAMP | Data de atualização |
| deleted_at | TIMESTAMP | Soft delete |

### 6. prescription_items

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| association_id | UUID | FK para associations |
| prescription_id | UUID | FK para prescriptions |
| product_id | UUID | FK para products |
| dosage_instructions | TEXT | Instruções de dosagem |
| quantity_recommended | INTEGER | Quantidade recomendada |
| created_at | TIMESTAMP | Data de criação |

### 7. products

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| association_id | UUID | FK para associations |
| name | VARCHAR(200) | Nome do produto |
| description | TEXT | Descrição |
| unit_price | DECIMAL(10,2) | Preço unitário |
| min_stock_alert | INTEGER | Estoque mínimo |
| is_active | BOOLEAN | Ativo? |
| created_at | TIMESTAMP | Data de criação |
| updated_at | TIMESTAMP | Data de atualização |
| deleted_at | TIMESTAMP | Soft delete |

### 8. product_lots

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| association_id | UUID | FK para associations |
| product_id | UUID | FK para products |
| lot_number | VARCHAR(50) | Número do lote |
| expiration_date | DATE | Data de validade |
| current_quantity | INTEGER | Quantidade atual |
| initial_quantity | INTEGER | Quantidade inicial |
| supplier | VARCHAR(200) | Fornecedor |
| purchase_date | DATE | Data de compra |
| purchase_price | DECIMAL(10,2) | Preço de compra |
| received_by | UUID | FK para users |
| received_at | TIMESTAMP | Data de recebimento |
| created_at | TIMESTAMP | Data de criação |
| updated_at | TIMESTAMP | Data de atualização |
| deleted_at | TIMESTAMP | Soft delete |

### 9. orders

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| association_id | UUID | FK para associations |
| patient_id | UUID | FK para patients |
| prescription_id | UUID | FK para prescriptions |
| status | VARCHAR(20) | pendente, separado, dispensa, correio, entregue, cancelado |
| shipping_carrier | VARCHAR(100) | Transportadora |
| tracking_code | VARCHAR(100) | Código de rastreio |
| shipping_label_url | TEXT | URL da etiqueta |
| shipping_cost | DECIMAL(10,2) | Custo do frete |
| total_amount | DECIMAL(10,2) | Valor total |
| notes | TEXT | Observações |
| order_date | TIMESTAMP | Data do pedido |
| status_updated_at | TIMESTAMP | Última atualização de status |
| created_at | TIMESTAMP | Data de criação |
| updated_at | TIMESTAMP | Data de atualização |
| deleted_at | TIMESTAMP | Soft delete |

### 10. order_items

| Coluna | Tipo | Descrição |
|--------|------|-----------|
| id | UUID | Chave primária |
| association_id | UUID | FK para associations |
| order_id | UUID | FK para orders |
| product_lot_id | UUID | FK para product_lots |
| quantity | INTEGER | Quantidade |
| unit_price | DECIMAL(10,2) | Preço unitário |
| total_price | DECIMAL(10,2) | **GERADO** (quantity * unit_price) |
| created_at | TIMESTAMP | Data de criação |

## 📊 Views

### 1. vw_expired_prescriptions
**Descrição:** Receitas vencidas
**Uso:** Alertas para renovar receitas

### 2. vw_low_stock
**Descrição:** Produtos com estoque baixo
**Uso:** Alertas de compra

### 3. vw_overdue_subscriptions
**Descrição:** Anuidades em atraso
**Uso:** Cobranças

### 4. vw_patient_dashboard
**Descrição:** Resumo de pacientes por status
**Uso:** Dashboard

### 5. vw_top_doctors
**Descrição:** Médicos que mais prescrevem
**Uso:** Relatórios

### 6. vw_stock_summary
**Descrição:** Resumo de estoque por produto
**Uso:** Relatórios

## 🔧 Funções e Triggers

### Funções

| Função | Descrição |
|--------|-----------|
| `update_updated_at_column()` | Atualiza updated_at automaticamente |
| `check_patient_limit()` | Verifica limite de pacientes |
| `enforce_patient_limit()` | Trigger para impedir excesso |
| `update_prescription_status()` | Atualiza status das receitas |
| `process_order_stock()` | Baixa estoque ao finalizar pedido |

### Triggers

| Trigger | Tabela | Evento | Função |
|---------|--------|--------|--------|
| `enforce_patient_limit_before_insert` | patients | BEFORE INSERT | enforce_patient_limit |
| `update_*_updated_at` | Várias | BEFORE UPDATE | update_updated_at_column |

📄 03-backend.md
markdown

# 🔧 Backend - Go + Chi + GORM

## 📁 Estrutura de Pastas

cannacare-app-v3/
├── cmd/
│ └── api/
│ └── main.go # Ponto de entrada da aplicação
├── internal/
│ ├── config/
│ │ └── config.go # Carrega configurações do .env
│ ├── database/
│ │ ├── db.go # Conexão com PostgreSQL (GORM)
│ │ └── migrate.go # Migrações do banco
│ ├── handlers/
│ │ ├── auth_handler.go # Autenticação
│ │ ├── patient_handler.go # Pacientes
│ │ ├── doctor_handler.go # Médicos
│ │ ├── prescription_handler.go # Receitas
│ │ ├── anamnese_handler.go # Acolhimento
│ │ ├── product_handler.go # Produtos
│ │ ├── stock_handler.go # Estoque
│ │ ├── order_handler.go # Pedidos
│ │ ├── financial_handler.go # Financeiro
│ │ ├── dashboard_handler.go # Dashboard
│ │ └── document_handler.go # Documentos
│ ├── middleware/
│ │ ├── auth.go # Autenticação JWT
│ │ ├── permissions.go # Permissões por role
│ │ └── logger.go # Logging
│ ├── models/
│ │ ├── base.go # BaseModel (ID, CreatedAt, UpdatedAt, DeletedAt)
│ │ ├── association.go # Association (multi-tenant)
│ │ ├── user.go # User
│ │ ├── patient.go # Patient
│ │ ├── doctor.go # Doctor
│ │ ├── prescription.go # Prescription
│ │ ├── prescription_item.go # PrescriptionItem
│ │ ├── product.go # Product
│ │ ├── product_lot.go # ProductLot
│ │ ├── order.go # Order
│ │ ├── order_item.go # OrderItem
│ │ ├── payment.go # Payment
│ │ ├── subscription.go # Subscription
│ │ ├── anamnese.go # Anamnese
│ │ ├── notification.go # Notification
│ │ └── patient_document.go # PatientDocument
│ ├── services/
│ │ ├── auth_service.go # Autenticação
│ │ ├── patient_service.go # Pacientes
│ │ ├── doctor_service.go # Médicos
│ │ ├── prescription_service.go # Receitas
│ │ ├── anamnese_service.go # Acolhimento
│ │ ├── product_service.go # Produtos
│ │ ├── stock_service.go # Estoque
│ │ ├── order_service.go # Pedidos
│ │ ├── financial_service.go # Financeiro
│ │ └── dashboard_service.go # Dashboard
│ └── utils/
│ ├── response.go # Padronização de respostas
│ └── validators.go # Validações (CPF, email, telefone)
├── pkg/
│ └── jwt/
│ └── jwt.go # JWT Service
├── uploads/
│ └── documents/ # Arquivos enviados
├── .env # Variáveis de ambiente
├── go.mod # Dependências
├── go.sum # Checksums
└── docker-compose.yml # Docker Compose
text


## 📦 Dependências

```go
// go.mod
module cannacare-backend

go 1.21

require (
    github.com/go-chi/chi/v5 v5.3.1          // HTTP router
    github.com/go-chi/cors v1.2.2            // CORS
    github.com/go-playground/validator/v10 v10.30.3 // Validação
    github.com/golang-jwt/jwt/v5 v5.3.1      // JWT
    github.com/google/uuid v1.6.0            // UUID
    github.com/joho/godotenv v1.5.1          // .env
    golang.org/x/crypto v0.54.0              // Bcrypt
    gorm.io/driver/postgres v1.6.0           // PostgreSQL driver
    gorm.io/gorm v1.31.2                     // ORM
)

🔧 Configuração (.env)
env

# ================================================================
# BANCO DE DADOS
# ================================================================
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=cannacare2026!
DB_NAME=cannacare_db
DB_SSLMODE=disable

# ================================================================
# SERVIDOR
# ================================================================
SERVER_PORT=8080

# ================================================================
# JWT (Autenticação)
# ================================================================
JWT_SECRET=cannacare-super-secret-key-2026
JWT_EXPIRES_IN=24h

# ================================================================
# AMBIENTE
# ================================================================
ENV=development

🚀 Como Rodar
bash

# 1. Subir o banco de dados (Docker)
docker compose up -d

# 2. Rodar o backend
go run cmd/api/main.go

# 3. Build para produção
go build -o cannacare ./cmd/api/main.go

📋 Endpoints Principais
Método	Endpoint	Descrição
POST	/api/auth/register	Registrar associação + admin
POST	/api/auth/login	Login
POST	/api/patients	Criar paciente
GET	/api/patients	Listar pacientes
GET	/api/patients/{id}	Buscar paciente
PUT	/api/patients/{id}	Atualizar paciente
DELETE	/api/patients/{id}	Remover paciente
POST	/api/doctors	Criar médico
GET	/api/doctors	Listar médicos
POST	/api/prescriptions	Criar receita
GET	/api/prescriptions	Listar receitas
POST	/api/orders	Criar pedido
GET	/api/orders	Listar pedidos
POST	/api/stock/lots	Adicionar lote
GET	/api/stock/lots	Listar lotes
POST	/api/financial/subscriptions	Criar anuidade
GET	/api/financial/subscriptions	Listar anuidades
GET	/api/dashboard/overview	Dashboard
🐳 Docker Compose
yaml

version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: cannacare_postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: cannacare2026!
      POSTGRES_DB: cannacare_db
    ports:
      - "5433:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: always

volumes:
  postgres_data:

🔐 Multi-Tenancy (Backend)
Como o association_id é passado

    JWT contém association_id

go

type Claims struct {
    UserID        string `json:"user_id"`
    AssociationID string `json:"association_id"` // ← ESSENCIAL!
    Email         string `json:"email"`
    Role          string `json:"role"`
    // ...
}

    Middleware extrai association_id

go

func AuthMiddleware(jwtService *jwt.JWTService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Validar token...
            claims, _ := jwtService.ValidateToken(token)
            
            // Adicionar ao Context
            ctx := context.WithValue(r.Context(), AssociationIDKey, claims.AssociationID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

    Handlers extraem do Context

go

func (h *PatientHandler) List(w http.ResponseWriter, r *http.Request) {
    associationID := r.Context().Value(middleware.AssociationIDKey).(string)
    // ...
}

    Services filtram por association_id

go

func (s *PatientService) List(associationID uuid.UUID, req ListPatientRequest) {
    query := s.db.Model(&models.Patient{}).Where("association_id = ?", associationID)
    // ...
}

text


---

## 📄 04-frontend.md

```markdown
# 🎨 Frontend - Next.js + TypeScript

## 📁 Estrutura de Pastas

cannacare-app-V3-Front/
└── cannacare-frontend/
├── app/
│ ├── dashboard/
│ │ ├── admin/
│ │ │ └── users/
│ │ │ └── page.tsx # Gerenciar usuários
│ │ ├── anamnesis/
│ │ │ ├── new/
│ │ │ │ └── page.tsx # Nova anamnese
│ │ │ └── page.tsx # Listar anamneses
│ │ ├── doctors/
│ │ │ └── page.tsx # Gerenciar médicos
│ │ ├── financial/
│ │ │ ├── payments/
│ │ │ │ └── page.tsx # Pagamentos
│ │ │ ├── subscriptions/
│ │ │ │ └── page.tsx # Anuidades
│ │ │ └── page.tsx # Financeiro
│ │ ├── orders/
│ │ │ ├── [id]/
│ │ │ │ └── page.tsx # Detalhe do pedido
│ │ │ ├── new/
│ │ │ │ └── page.tsx # Novo pedido
│ │ │ └── page.tsx # Listar pedidos
│ │ ├── patients/
│ │ │ ├── [id]/
│ │ │ │ ├── anamnesis/
│ │ │ │ │ ├── new/
│ │ │ │ │ │ └── page.tsx
│ │ │ │ │ └── page.tsx
│ │ │ │ ├── documents/
│ │ │ │ │ └── page.tsx # Documentos do paciente
│ │ │ │ ├── edit/
│ │ │ │ │ └── page.tsx # Editar paciente
│ │ │ │ └── page.tsx # Detalhe do paciente
│ │ │ ├── new/
│ │ │ │ └── page.tsx # Novo paciente
│ │ │ └── page.tsx # Listar pacientes
│ │ ├── prescriptions/
│ │ │ ├── new/
│ │ │ │ └── page.tsx # Nova receita
│ │ │ └── page.tsx # Listar receitas
│ │ ├── products/
│ │ │ └── page.tsx # Gerenciar produtos
│ │ ├── profile/
│ │ │ └── page.tsx # Perfil do usuário
│ │ ├── reports/
│ │ │ └── page.tsx # Relatórios
│ │ ├── stock/
│ │ │ ├── lots/
│ │ │ │ └── page.tsx # Gerenciar lotes
│ │ │ ├── movements/
│ │ │ │ └── page.tsx # Movimentações
│ │ │ └── page.tsx # Estoque
│ │ ├── layout.tsx # Layout do dashboard
│ │ └── page.tsx # Dashboard
│ ├── login/
│ │ └── page.tsx # Login
│ ├── register/
│ │ └── page.tsx # Registro
│ ├── globals.css # Estilos globais
│ └── layout.tsx # Layout principal
├── components/
│ ├── forms/
│ │ ├── DoctorForm.tsx # Formulário de médico
│ │ └── DocumentUpload.tsx # Upload de documentos
│ ├── layout/
│ │ ├── Header.tsx # Cabeçalho
│ │ └── Sidebar.tsx # Menu lateral
│ └── ui/
│ ├── Button.tsx # Componente Button
│ └── Card.tsx # Componente Card
├── lib/
│ └── api/
│ ├── client.ts # Axios client
│ ├── auth.ts # Autenticação
│ ├── patients.ts # Pacientes
│ ├── doctors.ts # Médicos
│ ├── prescriptions.ts # Receitas
│ ├── orders.ts # Pedidos
│ ├── stock.ts # Estoque
│ ├── financial.ts # Financeiro
│ └── dashboard.ts # Dashboard
├── middleware.ts # Middleware Next.js
├── package.json # Dependências
└── tailwind.config.ts # Configuração Tailwind
text


## 📦 Dependências

```json
{
  "dependencies": {
    "next": "14.2.4",
    "react": "18.2.0",
    "react-dom": "18.2.0",
    "axios": "^1.7.2",
    "react-hook-form": "^7.51.5",
    "@hookform/resolvers": "^3.6.0",
    "zod": "^3.23.8",
    "lucide-react": "^0.399.0"
  },
  "devDependencies": {
    "@types/node": "^20",
    "@types/react": "^18",
    "@types/react-dom": "^18",
    "tailwindcss": "^3.4.1",
    "autoprefixer": "^10",
    "postcss": "^8",
    "typescript": "^5"
  }
}

🔐 Autenticação (Frontend)
Login Flow
typescript

// lib/api/auth.ts
export async function login(data: LoginData): Promise<AuthResponse> {
    const response = await api.post("/api/auth/login", data);
    return response.data;
}

// app/login/page.tsx
const handleSubmit = async (e: React.FormEvent) => {
    const response = await login({ email, password });
    if (response.success) {
        // Salvar token
        localStorage.setItem("token", response.data.token);
        localStorage.setItem("user", JSON.stringify(response.data.user));
        // Redirecionar
        router.push("/dashboard");
    }
};

Interceptor (Client)
typescript

// lib/api/client.ts
const api = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem("token");
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

Middleware (Next.js)
typescript

// middleware.ts
export function middleware(request: NextRequest) {
    const token = request.cookies.get("token")?.value;
    const isPublicRoute = ["/login", "/register"].includes(pathname);
    
    if (!token && !isPublicRoute) {
        return NextResponse.redirect(new URL("/login", request.url));
    }
    return NextResponse.next();
}

🎨 Estilização (Tailwind CSS)
css

/* app/globals.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

/* Cores personalizadas */
:root {
    --primary: #2d6a4f;
    --secondary: #52b788;
    --dark: #1a3a2a;
    --cream: #f8f9fa;
}

📱 Páginas Principais
Dashboard

    Visão geral com cards estatísticos

    Fila Regulatória (pacientes em análise)

    Ações rápidas

Pacientes (Fila Regulatória)

    Lista de pacientes com status

    Mudança de status via dropdown

    Aprovação/rejeição de pacientes

Médicos

    Cadastro com CRM e especialidade

    Lista com ações (editar, ativar/desativar)

Receitas

    Criar com múltiplos itens

    Listar com validação

    Upload de arquivo (receita scaneada)

Acolhimento (Anamnese)

    Criar anamnese por paciente

    Histórico de anamneses

Estoque

    Produtos (CRUD)

    Lotes com validade

    Movimentações

Pedidos

    Criar com validação de receita

    Atualizar status

    Código de rastreio

Financeiro

    Anuidades

    Pagamentos

    Anuidades em atraso

Relatórios

    Pacientes por status

    Receitas vencidas

    Médicos top

    Estoque baixo

text


---

## 📄 05-multi-tenancy.md

```markdown
# 🔒 Multi-Tenancy - Isolamento de Dados

## 📋 O que é Multi-Tenancy?

Multi-tenancy é uma arquitetura onde **uma única instância do software** serve **múltiplos clientes** (associações), mantendo os dados de cada um **isolados**.

## 🏗️ Estratégia Adotada

O CannaCare utiliza a estratégia de **filtro por association_id**:

Todas as tabelas têm uma coluna association_id.
Todas as queries SQL filtram por association_id.
text


## 📊 Como Funciona

### 1. Tabela Mestra: associations

```sql
CREATE TABLE associations (
    id UUID PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    cnpj VARCHAR(18) UNIQUE NOT NULL,
    email VARCHAR(200) UNIQUE NOT NULL,
    plan VARCHAR(50) DEFAULT 'basic',
    status VARCHAR(20) DEFAULT 'pending',
    patient_limit INTEGER DEFAULT 50,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

2. Todas as Tabelas Têm association_id
sql

CREATE TABLE patients (
    id UUID PRIMARY KEY,
    association_id UUID NOT NULL REFERENCES associations(id),
    full_name VARCHAR(200) NOT NULL,
    cpf VARCHAR(14) NOT NULL,
    -- ...
    UNIQUE(cpf, association_id)
);

CREATE TABLE doctors (
    id UUID PRIMARY KEY,
    association_id UUID NOT NULL REFERENCES associations(id),
    name VARCHAR(200) NOT NULL,
    -- ...
    UNIQUE(crm, crm_state, association_id)
);

3. JWT Contém association_id
go

type Claims struct {
    UserID        string `json:"user_id"`
    AssociationID string `json:"association_id"` // ← ESSENCIAL!
    Email         string `json:"email"`
    Role          string `json:"role"`
    // ...
}

4. Middleware Extrai association_id
go

func AuthMiddleware(jwtService *jwt.JWTService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, _ := jwtService.ValidateToken(token)
            ctx := context.WithValue(r.Context(), AssociationIDKey, claims.AssociationID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

5. Todas as Queries Filtram por association_id
go

func (s *PatientService) List(associationID uuid.UUID) ([]Patient, error) {
    // ⚠️ SEMPRE filtrar por association_id!
    var patients []Patient
    err := s.db.Where("association_id = ?", associationID).Find(&patients).Error
    return patients, err
}

6. Constraints de Unicidade por Associação
sql

-- Email único por associação
UNIQUE(email, association_id)

-- CPF único por associação
UNIQUE(cpf, association_id)

-- CRM único por associação
UNIQUE(crm, crm_state, association_id)

🔐 Isolamento em Ação
Exemplo: Duas Associações
sql

-- Associação A (id: 8e0b...)
-- Associação B (id: 9f1c...)

-- Tabela patients:
-- Paciente 1: association_id = 8e0b... (Associação A)
-- Paciente 2: association_id = 8e0b... (Associação A)
-- Paciente 3: association_id = 9f1c... (Associação B)

-- Query da Associação A:
SELECT * FROM patients WHERE association_id = '8e0b...'
-- Retorna: Paciente 1, Paciente 2

-- Query da Associação B:
SELECT * FROM patients WHERE association_id = '9f1c...'
-- Retorna: Paciente 3

📈 Benefícios
Benefício	Descrição
Isolamento	Dados de cada associação são isolados
Segurança	Usuário só vê dados da sua associação
Eficiência	Uma instância atende todos os clientes
Custo	Menos infraestrutura para manter
Manutenção	Atualizações centralizadas
🚀 Onboarding de Nova Associação
bash

# 1. Associação se registra
POST /api/auth/register
{
    "association_name": "Associação XYZ",
    "cnpj": "12.345.678/0001-99",
    "email": "admin@xyz.com",
    "password": "123456"
}

# 2. Sistema cria:
# - Association ID: 9f1c...
# - Usuário admin vinculado à associação

# 3. Associação faz login
POST /api/auth/login
{
    "email": "admin@xyz.com",
    "password": "123456"
}

# 4. JWT retorna com association_id
{
    "token": "eyJ...",
    "user": {
        "association_id": "9f1c..."
    }
}

# 5. Todas as requisições futuras usam este association_id

🛡️ Segurança
Camada	Medida
Banco	Índices em association_id
API	Middleware extrai association_id do JWT
Queries	Todas as queries filtram por association_id
Constraints	UNIQUE por associação (email, CPF, CRM)
Soft Delete	DeletedAt mantém histórico
text


---

## 📄 06-autenticacao.md

```markdown
# 🔐 Autenticação - JWT

## 📋 Visão Geral

O CannaCare utiliza **JWT (JSON Web Token)** para autenticação. O token contém as informações do usuário e é assinado digitalmente.

## 🔑 Estrutura do JWT

### Header
```json
{
    "alg": "HS256",
    "typ": "JWT"
}

Payload (Claims)
json

{
    "user_id": "ed7c8ffc-9a7a-44c1-93c4-3afed1eed956",
    "association_id": "35df6913-623c-4bba-9815-3392d3995b4e",
    "email": "admin@cannacare.com",
    "name": "Administrador",
    "role": "admin",
    "exp": 1735689600,
    "iat": 1735686000
}

Signature
text

HS256(
    base64UrlEncode(header) + "." +
    base64UrlEncode(payload),
    secret
)

🔧 Implementação
JWT Service
go

// pkg/jwt/jwt.go
type Claims struct {
    UserID        string `json:"user_id"`
    AssociationID string `json:"association_id"` // ← ESSENCIAL!
    Email         string `json:"email"`
    Name          string `json:"name"`
    Role          string `json:"role"`
    jwt.RegisteredClaims
}

type JWTService struct {
    secretKey string
    expiresIn time.Duration
}

func (s *JWTService) GenerateToken(userID, associationID uuid.UUID, email, name, role string) (string, error) {
    claims := Claims{
        UserID:        userID.String(),
        AssociationID: associationID.String(),
        Email:         email,
        Name:          name,
        Role:          role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiresIn)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.secretKey))
}

Login Flow
go

// internal/services/auth_service.go
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
    // 1. Buscar usuário
    var user models.User
    s.db.Where("email = ?", req.Email).First(&user)
    
    // 2. Verificar senha (bcrypt)
    bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
    
    // 3. Gerar token
    token, _ := s.jwtService.GenerateToken(
        user.ID,
        user.AssociationID, // ← association_id vai no JWT!
        user.Email,
        user.Name,
        user.Role,
    )
    
    // 4. Atualizar last_login_at
    now := time.Now()
    user.LastLoginAt = &now
    s.db.Save(&user)
    
    // 5. Retornar token
    return &AuthResponse{
        Token:     token,
        TokenType: "Bearer",
        ExpiresIn: 86400,
        User:      toUserResponse(&user),
    }, nil
}

🔐 Middleware de Autenticação
go

// internal/middleware/auth.go
func AuthMiddleware(jwtService *jwt.JWTService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Extrair token do header
            authHeader := r.Header.Get("Authorization")
            parts := strings.Split(authHeader, " ")
            tokenString := parts[1]
            
            // 2. Validar token
            claims, err := jwtService.ValidateToken(tokenString)
            if err != nil {
                utils.SendError(w, http.StatusUnauthorized, "token inválido")
                return
            }
            
            // 3. Adicionar ao Context
            ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
            ctx = context.WithValue(ctx, AssociationIDKey, claims.AssociationID)
            ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
            ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
            ctx = context.WithValue(ctx, UserNameKey, claims.Name)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

🔒 Rotas Protegidas
go

// cmd/api/main.go
r.Group(func(r chi.Router) {
    // ⚠️ Aplica autenticação em todas as rotas abaixo
    r.Use(middleware.AuthMiddleware(jwtService))
    
    // Rotas protegidas
    r.Get("/api/protected", protectedHandler)
    r.Post("/api/patients", patientHandler.Create)
    r.Get("/api/patients", patientHandler.List)
    // ...
})

👑 Permissões (Roles)
go

// internal/middleware/permissions.go
var RolePermissions = map[string][]string{
    "admin":        {"*"},
    "coordenacao":  {"patients:read", "patients:write", "reports:read", "dashboard:read"},
    "secretaria":   {"patients:read", "patients:write", "documents:upload", "payments:read"},
    "acolhimento":  {"patients:read", "anamnesis:write"},
    "farmacia":     {"orders:read", "orders:write", "stock:read", "stock:write"},
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            role := r.Context().Value(UserRoleKey).(string)
            for _, allowed := range allowedRoles {
                if role == allowed {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            utils.SendError(w, http.StatusForbidden, "permissão insuficiente")
        })
    }
}

📋 Uso no Frontend
typescript

// 1. Login
const response = await login({ email, password });
localStorage.setItem("token", response.data.token);
localStorage.setItem("user", JSON.stringify(response.data.user));

// 2. Enviar token em todas as requisições
api.interceptors.request.use((config) => {
    const token = localStorage.getItem("token");
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// 3. Middleware (Next.js)
export function middleware(request: NextRequest) {
    const token = request.cookies.get("token")?.value;
    if (!token && !isPublicRoute) {
        return NextResponse.redirect(new URL("/login", request.url));
    }
    return NextResponse.next();
}

🔑 Credenciais Padrão
Email	Senha	Role
admin@associacao.com	123456	admin
admin@cannacare.com	admin123	admin
🛡️ Segurança
Aspecto	Medida
Senhas	Bcrypt (hash + salt)
Token	JWT com assinatura HS256
Expiração	24 horas
Transporte	HTTPS (produção)
Armazenamento	localStorage (frontend)
Renovação	Novo login após expirar
text


---

## 📄 07-rotas-api.md

```markdown
# 📡 Rotas da API - CannaCare

## 🔐 Autenticação

### POST /api/auth/register
Registra uma nova associação e cria o usuário admin.

**Request:**
```json
{
    "association_name": "Associação XYZ",
    "cnpj": "12.345.678/0001-99",
    "phone": "(11) 99999-9999",
    "name": "João Silva",
    "email": "admin@xyz.com",
    "password": "123456"
}

Response (201 Created):
json

{
    "success": true,
    "data": {
        "id": "abc-123",
        "association_id": "550e8400-...",
        "name": "João Silva",
        "email": "admin@xyz.com",
        "role": "admin",
        "is_active": true,
        "created_at": "2026-07-30T14:26:40Z"
    }
}

POST /api/auth/login

Autentica um usuário.

Request:
json

{
    "email": "admin@xyz.com",
    "password": "123456"
}

Response (200 OK):
json

{
    "success": true,
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "token_type": "Bearer",
        "expires_in": 86400,
        "user": {
            "id": "abc-123",
            "association_id": "550e8400-...",
            "name": "João Silva",
            "email": "admin@xyz.com",
            "role": "admin",
            "is_active": true,
            "last_login_at": "2026-07-30T14:26:48Z",
            "created_at": "2026-07-30T14:26:40Z"
        }
    }
}

👤 Pacientes
POST /api/patients

Cria um novo paciente.

Headers: Authorization: Bearer <token>

Request:
json

{
    "full_name": "Maria Oliveira",
    "birth_date": "1985-05-15T00:00:00Z",
    "gender": "Feminino",
    "cpf": "12345678901",
    "rg": "123456789",
    "phone": "(11) 98888-8888",
    "whatsapp": "(11) 98888-8888",
    "email": "maria@email.com",
    "address_street": "Rua das Flores",
    "address_number": "123",
    "address_neighborhood": "Centro",
    "address_city": "São Paulo",
    "address_state": "SP",
    "address_zipcode": "01000-000"
}

Response (201 Created):
json

{
    "success": true,
    "data": {
        "id": "patient-123",
        "association_id": "550e8400-...",
        "full_name": "Maria Oliveira",
        "cpf": "12345678901",
        "status": "pendente_documentacao",
        "created_at": "2026-07-30T14:30:00Z"
    }
}

GET /api/patients

Lista pacientes da associação.

Headers: Authorization: Bearer <token>

Query Params:

    name (string) - Filtrar por nome

    cpf (string) - Filtrar por CPF

    status (string) - Filtrar por status

    page (int) - Página (default: 1)

    limit (int) - Limite (default: 20)

Response (200 OK):
json

{
    "success": true,
    "data": {
        "items": [
            {
                "id": "patient-123",
                "full_name": "Maria Oliveira",
                "cpf": "12345678901",
                "status": "pendente_documentacao",
                "created_at": "2026-07-30T14:30:00Z"
            }
        ],
        "total": 1,
        "page": 1,
        "limit": 20
    }
}

GET /api/patients/{id}

Busca um paciente por ID.

Headers: Authorization: Bearer <token>

Response (200 OK):
json

{
    "success": true,
    "data": {
        "id": "patient-123",
        "association_id": "550e8400-...",
        "full_name": "Maria Oliveira",
        "birth_date": "1985-05-15",
        "gender": "Feminino",
        "cpf": "12345678901",
        "phone": "(11) 98888-8888",
        "email": "maria@email.com",
        "status": "pendente_documentacao",
        "is_social_patient": false,
        "created_at": "2026-07-30T14:30:00Z"
    }
}

PUT /api/patients/{id}

Atualiza um paciente.

Headers: Authorization: Bearer <token>

Request: (campos opcionais)
json

{
    "full_name": "Maria Oliveira Silva",
    "phone": "(11) 97777-7777",
    "email": "maria.silva@email.com"
}

Response (200 OK):
json

{
    "success": true,
    "data": { ... }
}

PATCH /api/patients/{id}/status

Atualiza o status de um paciente.

Headers: Authorization: Bearer <token>

Request:
json

{
    "status": "aprovado",
    "reason": "Documentos aprovados"
}

Response (200 OK):
json

{
    "success": true,
    "data": { ... }
}

DELETE /api/patients/{id}

Remove um paciente (soft delete).

Headers: Authorization: Bearer <token>

Response (200 OK):
json

{
    "success": true,
    "data": {
        "message": "Paciente removido com sucesso"
    }
}

GET /api/patients/stats

Estatísticas de pacientes.

Headers: Authorization: Bearer <token>

Response (200 OK):
json

{
    "success": true,
    "data": {
        "statuses": [
            { "status": "aprovado", "total": 10 },
            { "status": "pendente_documentacao", "total": 5 }
        ]
    }
}

👨‍⚕️ Médicos
POST /api/doctors

Cria um médico.

Headers: Authorization: Bearer <token>

Request:
json

{
    "name": "Dr. Carlos Silva",
    "crm": "12345",
    "crm_state": "SP",
    "specialty": "Neurologia",
    "phone": "(11) 99999-9999",
    "email": "carlos@medico.com"
}

Response (201 Created):
json

{
    "success": true,
    "data": {
        "id": "doctor-123",
        "name": "Dr. Carlos Silva",
        "crm": "12345",
        "crm_state": "SP",
        "specialty": "Neurologia",
        "is_active": true,
        "created_at": "2026-07-30T14:30:00Z"
    }
}

GET /api/doctors

Lista médicos.

Headers: Authorization: Bearer <token>

Query Params:

    name (string) - Filtrar por nome

    specialty (string) - Filtrar por especialidade

    is_active (boolean) - Filtrar por status

Response (200 OK):
json

{
    "success": true,
    "data": {
        "items": [ ... ],
        "total": 1,
        "page": 1,
        "limit": 20
    }
}

GET /api/doctors/top

Médicos que mais prescrevem.

Headers: Authorization: Bearer <token>

Response (200 OK):
json

{
    "success": true,
    "data": [
        {
            "doctor_name": "Dr. Carlos Silva",
            "total_prescriptions": 45,
            "unique_patients": 30
        }
    ]
}

💊 Receitas
POST /api/prescriptions

Cria uma receita.

Headers: Authorization: Bearer <token>

Request:
json

{
    "patient_id": "patient-123",
    "doctor_id": "doctor-123",
    "cid": "G40.0",
    "issue_date": "2026-07-30T00:00:00Z",
    "expiration_date": "2027-07-30T00:00:00Z",
    "items": [
        {
            "product_id": "product-123",
            "dosage_instructions": "5 gotas a cada 12 horas",
            "quantity_recommended": 2
        }
    ]
}

Response (201 Created):
json

{
    "success": true,
    "data": {
        "id": "prescription-123",
        "status": "valida",
        "is_active": true,
        "days_until_expire": 365,
        "items": [ ... ]
    }
}

GET /api/prescriptions

Lista receitas.

Headers: Authorization: Bearer <token>

Query Params:

    patient_id (string) - Filtrar por paciente

    status (string) - Filtrar por status

    is_active (boolean) - Filtrar por ativo

Response (200 OK):
json

{
    "success": true,
    "data": {
        "items": [ ... ],
        "total": 1,
        "page": 1,
        "limit": 20
    }
}

GET /api/prescriptions/validate/{id}

Valida uma receita.

Headers: Authorization: Bearer <token>

Response (200 OK):
json

{
    "success": true,
    "data": {
        "is_valid": true,
        "message": "prescrição válida"
    }
}

📦 Estoque
POST /api/stock/lots

Adiciona um lote.

Headers: Authorization: Bearer <token>

Request:
json

{
    "product_id": "product-123",
    "lot_number": "LOTE-001",
    "expiration_date": "2027-12-31T00:00:00Z",
    "quantity": 100,
    "supplier": "Fornecedor XYZ",
    "purchase_date": "2026-07-30T00:00:00Z",
    "purchase_price": 50.00
}

Response (201 Created):
json

{
    "success": true,
    "data": {
        "id": "lot-123",
        "product_name": "Óleo CBD 10%",
        "lot_number": "LOTE-001",
        "current_quantity": 100,
        "is_expired": false,
        "days_until_expire": 519
    }
}

GET /api/stock/lots

Lista lotes.

Headers: Authorization: Bearer <token>

Query Params:

    product_id (string) - Filtrar por produto

    is_expired (boolean) - Filtrar vencidos

Response (200 OK):
json

{
    "success": true,
    "data": {
        "items": [ ... ],
        "total": 1,
        "page": 1,
        "limit": 20
    }
}

🛒 Pedidos
POST /api/orders

Cria um pedido.

Headers: Authorization: Bearer <token>

Request:
json

{
    "patient_id": "patient-123",
    "prescription_id": "prescription-123",
    "items": [
        {
            "product_lot_id": "lot-123",
            "quantity": 2,
            "unit_price": 100.00
        }
    ],
    "notes": "Pedido urgente"
}

Response (201 Created):
json

{
    "success": true,
    "data": {
        "id": "order-123",
        "status": "pendente",
        "total_amount": 200.00,
        "order_date": "2026-07-30T14:30:00Z"
    }
}

PATCH /api/orders/{id}/status

Atualiza status do pedido.

Headers: Authorization: Bearer <token>

Request:
json

{
    "status": "separado",
    "notes": "Produto separado"
}

Response (200 OK):
json

{
    "success": true,
    "data": { ... }
}

💰 Financeiro
POST /api/financial/subscriptions

Cria uma anuidade.

Headers: Authorization: Bearer <token>

Request:
json

{
    "patient_id": "patient-123",
    "due_date": "2027-07-30T00:00:00Z",
    "amount": 150.00
}

Response (201 Created):
json

{
    "success": true,
    "data": {
        "id": "subscription-123",
        "patient_name": "Maria Oliveira",
        "due_date": "2027-07-30",
        "amount": 150.00,
        "status": "pendente"
    }
}

POST /api/financial/payments

Registra um pagamento.

Headers: Authorization: Bearer <token>

Request:
json

{
    "patient_id": "patient-123",
    "subscription_id": "subscription-123",
    "payment_type": "anuidade",
    "payment_method": "pix",
    "amount": 150.00,
    "status": "pago"
}

Response (201 Created):
json

{
    "success": true,
    "data": {
        "id": "payment-123",
        "patient_name": "Maria Oliveira",
        "payment_type": "anuidade",
        "payment_method": "pix",
        "amount": 150.00,
        "status": "pago"
    }
}

📊 Dashboard
GET /api/dashboard/overview

Visão geral do sistema.

Headers: Authorization: Bearer <token>

Response (200 OK):
json

{
    "success": true,
    "data": {
        "patients": {
            "total": 15,
            "approved": 10,
            "pending": 5
        },
        "doctors": {
            "total": 3,
            "active": 3
        },
        "orders": {
            "total": 8,
            "pending": 2,
            "delivered": 6
        },
        "financial": {
            "total_revenue": 1500.00,
            "overdue_subscriptions": 2
        },
        "stock": {
            "total_quantity": 150,
            "low_stock_items": 3
        },
        "updated_at": "2026-07-30T14:30:00Z"
    }
}

📋 Códigos de Status
Código	Descrição
200	OK - Requisição bem-sucedida
201	Created - Recurso criado
400	Bad Request - Dados inválidos
401	Unauthorized - Não autenticado
403	Forbidden - Sem permissão
404	Not Found - Recurso não encontrado
409	Conflict - Conflito (ex: CPF duplicado)
500	Internal Server Error - Erro no servidor
text


---

## 📄 08-deploy.md

```markdown
# 🚀 Guia de Deploy

## 📋 Visão Geral

Este guia cobre o deploy do CannaCare em produção usando serviços de nuvem.

## 🏗️ Arquitetura de Deploy

┌─────────────────────────────────────────────────────────────────────────────┐
│ INTERNET │
└─────────────────────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ VERCEL (Frontend) │
│ • Next.js 14 │
│ • Domínio: cannacare.vercel.app ou domínio próprio │
│ • CDN Global │
└─────────────────────────────────────────────────────────────────────────────┘
│
▼ (API Calls)
┌─────────────────────────────────────────────────────────────────────────────┐
│ RENDER (Backend) │
│ • Go API │
│ • Domínio: api.cannacare.com │
│ • SSL automático │
└─────────────────────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ NEON (PostgreSQL) │
│ • Serverless │
│ • Escala automática │
│ • Branches para testes │
└─────────────────────────────────────────────────────────────────────────────┘
text


## 🟢 Backend: Render

### 1. Criar conta no Render

https://render.com
text


### 2. Conectar com GitHub

### 3. Criar Web Service

**Configurações:**
- **Name:** `cannacare-backend`
- **Environment:** `Docker`
- **Repository:** `GitHubAlves150/cannacare-app-v3`
- **Branch:** `main`
- **Dockerfile Path:** `./Dockerfile`

### 4. Variáveis de Ambiente

```env
DB_HOST=ep-cool-snow-123456.us-east-1.aws.neon.tech
DB_PORT=5432
DB_USER=cannacare
DB_PASSWORD=sua_senha_do_neon
DB_NAME=cannacare_db
DB_SSLMODE=require
JWT_SECRET=uma_chave_secreta_forte_aleatoria
JWT_EXPIRES_IN=24h
SERVER_PORT=8080
ENV=production

5. Dockerfile
dockerfile

FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]

⚡ Frontend: Vercel
1. Criar conta no Vercel
text

https://vercel.com

2. Conectar com GitHub
3. Importar Projeto

Configurações:

    Framework Preset: Next.js

    Root Directory: cannacare-frontend

4. Variáveis de Ambiente
env

NEXT_PUBLIC_API_URL=https://api.cannacare.com

🐘 Banco de Dados: Neon
1. Criar conta no Neon
text

https://neon.tech

2. Criar Projeto
3. Obter String de Conexão
text

postgresql://cannacare:senha@ep-cool-snow-123456.us-east-1.aws.neon.tech/cannacare_db?sslmode=require

4. Rodar Migrações
bash

# Via Render (automático) ou manual:
psql -h ep-cool-snow-123456.us-east-1.aws.neon.tech -U cannacare -d cannacare_db -f migrate.sql

🌐 Domínio Próprio
Configurar DNS
Registro	Tipo	Valor
api.cannacare.com	CNAME	render.com
cannacare.com	CNAME	vercel-dns.com
SSL

    Render: SSL automático

    Vercel: SSL automático

    Neon: SSL automático

🛠️ Docker Compose (Desenvolvimento)
yaml

version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: cannacare_postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: cannacare2026!
      POSTGRES_DB: cannacare_db
    ports:
      - "5433:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: always

volumes:
  postgres_data:

🔧 CI/CD (GitHub Actions)
yaml

name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to Render
        run: |
          curl -X POST https://api.render.com/deploy/srv-xxx?key=${{ secrets.RENDER_KEY }}

  deploy-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to Vercel
        run: |
          npm install -g vercel
          vercel --prod --token ${{ secrets.VERCEL_TOKEN }}

💰 Custos Estimados
Serviço	Plano	Custo	Inclui
Neon	Free Tier	$0	0.5GB storage, 10GB transfer
Render	Starter	$7/mês	Web Service + 256MB RAM
Vercel	Hobby	$0	100GB bandwidth, 100 deploys/dia
Total		$7/mês	
🚀 Comandos Rápidos
bash

# Build do backend
docker build -t cannacare-backend .

# Rodar localmente
docker run -p 8080:8080 --env-file .env cannacare-backend

# Deploy manual (Render)
git push origin main

# Deploy manual (Vercel)
cd cannacare-frontend
vercel --prod

🔒 Segurança em Produção
Item	Medida
Senhas	Usar senhas fortes (openssl rand -base64 32)
JWT	Chave secreta aleatória
CORS	Restringir para domínios específicos
HTTPS	Forçar HTTPS
Banco	SSL obrigatório
Logs	Não logar senhas ou tokens
text


---

## 📄 09-faq.md

```markdown
# ❓ Perguntas Frequentes (FAQ)

## 📋 Geral

### O que é multi-tenancy?
Multi-tenancy é uma arquitetura onde uma única instância do software serve múltiplos clientes, mantendo os dados de cada um isolados. No CannaCare, cada associação é um "tenant" com seus próprios dados isolados.

### Como os dados são isolados?
Todas as tabelas têm uma coluna `association_id`. O sistema sempre filtra por este ID em todas as queries, garantindo que cada associação veja apenas seus próprios dados.

### O sistema suporta quantas associações?
Ilimitadas. O multi-tenancy foi projetado para escalar horizontalmente.

## 🔐 Autenticação

### Como faço login?
Use as credenciais padrão:
- Email: `admin@associacao.com`
- Senha: `123456`

### Como criar um novo usuário?
1. Faça login como admin
2. Vá em Admin → Gerenciar Usuários
3. Clique em "Mudar função" para alterar permissões

### Esqueci minha senha, como recuperar?
Atualmente, contate o administrador do sistema para redefinir a senha.

### O token JWT expira?
Sim, após 24 horas. O usuário precisa fazer login novamente.

## 🗄️ Banco de Dados

### Qual banco de dados é usado?
PostgreSQL 15+ com suporte a UUID e JSONB.

### Como faço backup do banco?
```bash
docker exec -it cannacare_postgres pg_dump -U postgres cannacare_db > backup.sql

Como restaurar um backup?
bash

docker exec -i cannacare_postgres psql -U postgres -d cannacare_db < backup.sql

O que é soft delete?

Os registros não são removidos fisicamente. A coluna deleted_at é preenchida, e o GORM automaticamente esconde estes registros.
🚀 Deploy
Onde posso hospedar?

    Backend: Render, Railway, Fly.io

    Frontend: Vercel, Netlify

    Banco: Neon, Supabase, Render PostgreSQL

Quanto custa?

Com planos gratuitos, cerca de $7/mês para o backend. Frontend e banco (Neon) são gratuitos.
Preciso de CI/CD?

Não obrigatório, mas recomendado para automatizar deploys.
Como configuro um domínio próprio?

    Configure registros DNS (CNAME ou A)

    Adicione o domínio no Render e Vercel

    Aguarde a propagação DNS (até 48h)

🐛 Erros Comuns
"connection refused" ao rodar backend

O PostgreSQL não está rodando. Execute:
bash

docker compose up -d

"invalid token" no frontend

O token expirou ou é inválido. Faça login novamente.
"permission denied" ao acessar rota

O usuário não tem permissão para acessar esta rota. Verifique a role do usuário.
"duplicate key value violates unique constraint"

CPF, CNPJ ou email já cadastrado.
📚 Desenvolvimento
Como adicionar uma nova tabela?

    Criar model em internal/models/

    Adicionar ao AutoMigrate em internal/database/migrate.go

    Criar handler e service correspondentes

Como adicionar uma nova rota?

    Adicionar handler em internal/handlers/

    Registrar rota em cmd/api/main.go

    Adicionar serviço em internal/services/

Como testar a API?
bash

# Teste de login
curl -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"email":"admin@associacao.com","password":"123456"}'

# Teste protegido
curl -X GET http://localhost:8080/api/patients -H "Authorization: Bearer <token>"

Como rodar em modo debug?

O backend já mostra logs detalhados das queries SQL.
🔒 Segurança
Como mudar a senha padrão?
sql

UPDATE users SET password_hash = '$2a$10$NOVO_HASH' WHERE email = 'admin@associacao.com';

Como gerar um novo JWT secret?
bash

openssl rand -base64 32

O sistema tem proteção contra SQL Injection?

Sim. O GORM usa prepared statements, protegendo contra SQL Injection.
Os dados são criptografados?

As senhas são criptografadas com bcrypt. Dados sensíveis como CPF não são criptografados, mas o acesso é restrito por association_id.
text


---

## 📁 COMO SALVAR ESTA DOCUMENTAÇÃO

```bash
# 1. Criar a pasta de documentação
mkdir -p ~/Documents/cannacare-app-v3/docs

# 2. Criar os arquivos
cd ~/Documents/cannacare-app-v3/docs

# 3. Salvar cada seção em um arquivo .md
# (copie o conteúdo de cada seção para o arquivo correspondente)

# 4. Adicionar ao git
git add docs/
git commit -m "docs: adiciona documentação completa do projeto"
git push

📋 ESTRUTURA DOS ARQUIVOS
text

docs/
├── 00-README.md
├── 01-arquitetura.md
├── 02-banco-de-dados.md
├── 03-backend.md
├── 04-frontend.md
├── 05-multi-tenancy.md
├── 06-autenticacao.md
├── 07-rotas-api.md
├── 08-deploy.md
└── 09-faq.md

Documentação concluída! Agora você tem um guia completo para qualquer desenvolvedor entender e dar sequência ao projeto. 🌿🚀



```

## Observação Final!
Para o seu cenário de até **100 clientes** com dados médicos isolados no Postgres, as melhores opções combinam facilidade de deploy de contêineres com segurança lógica rigorosa.

Aqui estão os serviços ideais e como configurar o isolamento:

---

## 1. Onde Hospedar a Infraestrutura

Para não perder tempo gerenciando servidores puros (como AWS EC2), utilize plataformas **PaaS (Plataforma como Serviço)** focadas em contêineres. Elas escalam automaticamente conforme seus clientes crescem.

### 🟢 Para o Backend (Go)
* **Render (Web Services)** ou **Railway**: Você só precisa apontar para o seu repositório do GitHub onde está o `Dockerfile` do Go. Eles constroem a imagem e rodam o contêiner automaticamente.

### ⚡ Para o Frontend (Next.js)
* **Vercel**: É a plataforma oficial dos criadores do Next.js. O deploy é nativo, oferece performance excelente globalmente e possui a melhor integração do mercado (*Server-Side Rendering* e otimização de imagens).

### 🐘 Para o Banco de Dados (PostgreSQL)
* **Neon**: Um Postgres focado em nuvem e *serverless*. Ele escala o armazenamento de forma automática e permite criar *"branches"* se precisar testar atualizações.
* **Render PostgreSQL / Railway PostgreSQL**: Se preferir manter o banco na mesma plataforma do backend Go para reduzir a latência.
