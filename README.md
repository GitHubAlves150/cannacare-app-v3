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
## 📁 ETAPA 8: Implementação do sisitema de acolhimento completo com anamenese inicial e rastreamento periódicos.

Objetivo desta etapa:
    O Objetivo desta etapa é criar o sistema de negócio de toda a triagem com o paicente.

📁 ESTRUTURA QUE VAMOS CRIAR.

```bash
cannacare-app-v3/
├── internal/
│   ├── models/
│   │   └── anamnese.go                # ✅ Já existe
│   ├── services/
│   │   ├── auth_service.go            # ✅ Já existe
│   │   ├── doctor_service.go          # ✅ Já existe
│   │   ├── patient_service.go         # ✅ Já existe
│   │   ├── document_service.go        # ✅ Já existe
│   │   ├── prescription_service.go    # ✅ Já existe
│   │   └── anamnese_service.go        # 🆕 Lógica para anamnese
│   ├── handlers/
│   │   ├── auth_handler.go            # ✅ Já existe
│   │   ├── doctor_handler.go          # ✅ Já existe
│   │   ├── patient_handler.go         # ✅ Já existe
│   │   ├── document_handler.go        # ✅ Já existe
│   │   ├── prescription_handler.go    # ✅ Já existe
│   │   └── anamnese_handler.go        # 🆕 Endpoints para anamnese
│   ├── middleware/
│   │   └── auth.go                    # ✅ Já existe
│   └── utils/
│       ├── response.go                # ✅ Já existe
│       └── validators.go              # ✅ Já existe
├── pkg/
│   └── jwt/
│       └── jwt.go                     # ✅ Já existe
└── cmd/api/main.go                    # 🔄 Vamos atualizar
```

## ✅ TESTAR OS ENDPOINTS DE ANAMNESE.

1. Criar anamnese inicial para paciente
```bash
TOKEN="seu-token-aqui"
PATIENT_ID="id-do-paciente-aprovado"

curl -X POST "http://localhost:8080/api/patients/$PATIENT_ID/anamnesis" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "inicial",
    "symptoms": "Dor de cabeça, tontura, ansiedade",
    "symptom_intensity": 7,
    "side_effects": "Náusea leve, fadiga",
    "side_effect_intensity": 3,
    "treatment_adherence": "alta",
    "challenges": "Dificuldade de adaptação ao medicamento",
    "improvements": "Melhora da qualidade de sono",
    "additional_notes": "Paciente relata melhora significativa",
    "weight": 72.5,
    "blood_pressure": "120/80",
    "heart_rate": 72,
    "extra_responses": {
      "observacao_extra": "Paciente com bom humor"
    }
  }'
``` 

2. Listar anamneses do paciente
```bash
curl -X GET "http://localhost:8080/api/patients/$PATIENT_ID/anamnesis" \
  -H "Authorization: Bearer $TOKEN"
``` 

3. Criar rastreamento de 1 mês
```bash
curl -X POST "http://localhost:8080/api/patients/$PATIENT_ID/anamnesis" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "rastreio_1_mes",
    "symptoms": "Dor de cabeça leve, ansiedade controlada",
    "symptom_intensity": 3,
    "side_effects": "Nenhum",
    "side_effect_intensity": 0,
    "treatment_adherence": "alta",
    "challenges": "Nenhum",
    "improvements": "Melhora significativa dos sintomas",
    "additional_notes": "Paciente muito satisfeito",
    "weight": 70.0,
    "blood_pressure": "115/75",
    "heart_rate": 68
  }'
``` 


4. Buscar anamnese por ID
```bash
ANAMNESE_ID="id-da-anamnese-criada"

curl -X GET "http://localhost:8080/api/anamnesis/$ANAMNESE_ID" \
  -H "Authorization: Bearer $TOKEN"
``` 

xxxx
```bash

``` 

xxxx
```bash

``` 


xxxx
```bash

``` 


xxxx
```bash

``` 


xxxx
```bash

``` 


xxxx
```bash

``` 


