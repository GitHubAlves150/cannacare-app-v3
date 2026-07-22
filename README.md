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
## 📁 ETAPA 6: Implementação do sisitema de upload de documentos.

Objetivo desta etapa:

    Nesta etapa será feito  implementação do sistema de upload de documentos para pacientes. Esta etapa é crucial para o fluxo de aprovação de pacientes.


📁 ESTRUTURA QUE VAMOS CRIAR.

```bash
cannacare-app-v3/
├── internal/
│   ├── models/
│   │   └── patient_document.go     # ✅ Já existe
│   ├── services/
│   │   ├── auth_service.go         # ✅ Já existe
│   │   ├── doctor_service.go       # ✅ Já existe
│   │   ├── patient_service.go      # ✅ Já existe
│   │   └── document_service.go     # 🆕 Lógica para documentos
│   ├── handlers/
│   │   ├── auth_handler.go         # ✅ Já existe
│   │   ├── doctor_handler.go       # ✅ Já existe
│   │   ├── patient_handler.go      # ✅ Já existe
│   │   └── document_handler.go     # 🆕 Endpoints para documentos
│   ├── middleware/
│   │   └── auth.go                 # ✅ Já existe
│   └── utils/
│       ├── response.go             # ✅ Já existe
│       └── validators.go           # ✅ Já existe
├── uploads/                         # 🆕 Pasta para armazenar arquivos
│   └── documents/                   # 🆕 Subpasta para documentos
├── pkg/
│   └── jwt/
│       └── jwt.go                  # ✅ Já existe
└── cmd/api/main.go                 # 🔄 Vamos atualizar
```

## ✅ TESTAR OS ENDPOINTS DE DOCUMENTOS.

1. Fazer login (obter token)
``` bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cannacare.com","password":"admin123"}'
```
Resposta:
```bash
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMjkzNmRiM2UtY2FhMy00YjM0LTk5OWItMTViNDM0Y2EzMzE1IiwiZW1haWwiOiJhZG1pbkBjYW5uYWNhcmUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzg0ODE1NTIyLCJuYmYiOjE3ODQ3MjkxMjIsImlhdCI6MTc4NDcyOTEyMn0.vKoeW_ldIaciy3bC_8gRjR6vBgEaUwim1MxAxZqdtQY",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": "2936db3e-caa3-4b34-999b-15b434ca3315",
      "name": "Administrador",
      "email": "admin@cannacare.com",
      "role": "admin",
      "is_active": true,
      "last_login_at": "2026-07-22T11:05:22.713813338-03:00",
      "created_at": "2026-07-21T13:25:18.668715-03:00"
    }
  }
}
```
## 2. Upload de documento (RG/CPF).
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMjkzNmRiM2UtY2FhMy00YjM0LTk5OWItMTViNDM0Y2EzMzE1IiwiZW1haWwiOiJhZG1pbkBjYW5uYWNhcmUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzg0ODE1NTIyLCJuYmYiOjE3ODQ3MjkxMjIsImlhdCI6MTc4NDcyOTEyMn0.vKoeW_ldIaciy3bC_8gRjR6vBgEaUwim1MxAxZqdtQY"
PATIENT_ID="2fac4ddc-5c04-4c2b-aa40-151738ee9b94"

curl -X POST "http://localhost:8080/api/patients/$PATIENT_ID/documents" \
  -H "Authorization: Bearer $TOKEN" \
  -F "document_type=rg_cpf" \
  -F "file=@/caminho/para/seu/arquivo.pdf"

```

## Resposta.
![alt text](image-17.png)

![alt text](image-18.png)

## 3. Listar documentos do paciente.
```bash
curl -X GET "http://localhost:8080/api/patients/$PATIENT_ID/documents" \
  -H "Authorization: Bearer $TOKEN"
```

## Resposta.
![alt text](image-19.png)

## 4. Aprovar documento.
```bash
DOCUMENT_ID="id-do-documento"

curl -X PATCH "http://localhost:8080/api/documents/$DOCUMENT_ID/status" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "aprovado",
    "reason": "Documento válido e legível"
  }'

```

## Resposta.
![alt text](image-20.png)

![alt text](image-21.png)


## 5. Baixar documento.
```bash
curl -X GET "http://localhost:8080/api/documents/$DOCUMENT_ID/download" \
  -H "Authorization: Bearer $TOKEN" \
  --output documento.pdf
``` 
## Resposta - Após rodado o comando será baixado o documento com o nomo document.pdf para dentro da pasta raiz do projeto.
![alt text](image-22.png)
