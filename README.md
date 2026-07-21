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
## 📁 ETAPA 5: CRUD de pacientes.

Objetivo desta etapa:

    O objetivo desta etapa é criar o CRUD de pacientes para, poder criar pacientes, atualizar os dados do paciente, deletar o paciente e ler os dados do paicente.

 CRUD completo de pacientes

    ✅ Criar paciente

    ✅ Listar pacientes (com filtros)

    ✅ Buscar paciente por ID

    ✅ Atualizar paciente

    ✅ Deletar paciente (soft delete)

Validações específicas

    ✅ CPF único

    ✅ Data de nascimento

    ✅ Email válido

    ✅ Telefone

Status do paciente

    ✅ pendente_documentacao - Aguardando documentos

    ✅ em_analise - Em análise pela equipe

    ✅ aprovado - Paciente aprovado

    ✅ negado - Paciente reprovado

    ✅ assistente_social - Em análise social

Diferenciais

    ✅ Associação com usuário (para acesso ao portal)

    ✅ Role automática "paciente"

    ✅ Paciente social (isenção de anuidade)

    ____

## Endpoints que vamos criar:

| Método | Endpoint | Descrição |
| :---: | :--- | :--- |
| **POST** | `/api/patients` | Criar paciente |
| **GET** | `/api/patients` | Listar pacientes (filtros) |
| **GET** | `/api/patients/{id}` | Buscar paciente por ID |
| **PUT** | `/api/patients/{id}` | Atualizar paciente |
| **DELETE** | `/api/patients/{id}` | Deletar paciente |
| **PATCH** | `/api/patients/{id}/status` | Mudar status do paciente |
| **GET** | `/api/patients/{id}/documents` | Listar documentos do paciente |

## 📁 ESTRUTURA QUE VAMOS CRIAR.
```bash
cannacare-app-v3/
├── internal/
│   ├── models/
│   │   └── patient.go              # ✅ Já existe
│   ├── services/
│   │   ├── auth_service.go         # ✅ Já existe
│   │   ├── doctor_service.go       # ✅ Já existe
│   │   └── patient_service.go      # 🆕 Lógica de negócio para pacientes
│   ├── handlers/
│   │   ├── auth_handler.go         # ✅ Já existe
│   │   ├── doctor_handler.go       # ✅ Já existe
│   │   └── patient_handler.go      # 🆕 Endpoints HTTP para pacientes
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

## ✅ TESTAR OS ENDPOINTS DE PACIENTES.

1. Fazer login (obter token)
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cannacare.com","password":"admin123"}'
```

## 2. Criar paciente.
```bash
TOKEN="seu-token-aqui"

curl -X POST http://localhost:8080/api/patients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "João da Silva",
    "birth_date": "1985-05-15T00:00:00Z",
    "gender": "Masculino",
    "cpf": "12345678901",
    "phone": "(11) 99999-9999",
    "whatsapp": "(11) 98888-8888",
    "email": "joao.silva@email.com",
    "address_street": "Rua das Flores",
    "address_number": "123",
    "address_neighborhood": "Centro",
    "address_city": "São Paulo",
    "address_state": "SP",
    "address_zipcode": "01000-000"
  }'
```

## 3. Listar pacientes.
```bash
curl -X GET "http://localhost:8080/api/patients?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

## 4. Buscar paciente por ID.
```bash
curl -X GET "http://localhost:8080/api/patients/{id}" \
  -H "Authorization: Bearer $TOKEN"
```

## 5. Atualizar status do paciente.
``` bash
curl -X PATCH "http://localhost:8080/api/patients/{id}/status" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "em_analise",
    "reason": "Documentos recebidos, iniciando análise"
  }'
```
