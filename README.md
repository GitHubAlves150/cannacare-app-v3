```bash

🧪 Guia de Testes de Isolamento Multi-Tenancy - CannaCare
📋 Introdução

Este guia documenta o processo de teste de isolamento de dados (multi-tenancy) do sistema CannaCare. O objetivo é verificar que cada associação (cliente) vê APENAS os seus próprios dados, garantindo a segurança e privacidade das informações.
O que é Multi-Tenancy?

Multi-tenancy é uma arquitetura onde uma única instância do software serve múltiplos clientes (associações), mantendo os dados de cada um isolados. No CannaCare, isso é feito através da coluna association_id em todas as tabelas.
text

┌─────────────────────────────────────────────────────────────────┐
│                     CANNACARE (SaaS)                          │
├─────────────────────────────────────────────────────────────────┤
│  ┌───────────────────┐  ┌───────────────────┐                 │
│  │  Associação A     │  │  Associação B     │                 │
│  │  (SP)             │  │  (RJ)             │                 │
│  │  ID: f6f5...      │  │  ID: f123...     │                 │
│  ├───────────────────┤  ├───────────────────┤                 │
│  │ ✅ João Silva     │  │ ✅ Maria Souza    │                 │
│  │ ✅ Dr. Ana Paula  │  │ ✅ Dr. Roberto    │                 │
│  │ ❌ Maria Souza    │  │ ❌ João Silva     │                 │
│  │ ❌ Dr. Roberto    │  │ ❌ Dr. Ana Paula  │                 │
│  └───────────────────┘  └───────────────────┘                 │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  BANCO DE DADOS (Todas as tabelas têm association_id)   │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘

🛠️ Pré-requisitos

    Backend rodando em http://localhost:8080

    Banco PostgreSQL acessível via Docker

    curl e jq instalados

bash

# Verificar se os serviços estão rodando
docker ps | grep cannacare_postgres
curl -s http://localhost:8080/health | jq '.'

📝 Passo 1: Limpar Dados Anteriores (Opcional)

Para começar do zero, limpe todos os dados:
bash

# Conectar ao banco
docker exec -it cannacare_postgres psql -U postgres -d cannacare_db

# Executar os comandos SQL:

sql

-- ================================================================
-- LIMPEZA COMPLETA DOS DADOS
-- ================================================================

-- 1. Desabilitar triggers temporariamente
ALTER TABLE patients DISABLE TRIGGER enforce_patient_limit_before_insert;

-- 2. Limpar tabelas (ordem correta - primeiro as que têm FK)
DELETE FROM order_items;
DELETE FROM stock_movements;
DELETE FROM payments;
DELETE FROM subscriptions;
DELETE FROM orders;
DELETE FROM prescription_items;
DELETE FROM prescriptions;
DELETE FROM patient_documents;
DELETE FROM patient_status_history;
DELETE FROM anamneses;
DELETE FROM notifications;
DELETE FROM patients;
DELETE FROM doctors;
DELETE FROM users;
DELETE FROM associations;

-- 3. Resetar sequências
TRUNCATE TABLE associations, users, patients, doctors, prescriptions, 
           prescription_items, products, product_lots, orders, order_items,
           payments, subscriptions, anamneses, notifications, patient_documents,
           patient_status_history, stock_movements RESTART IDENTITY CASCADE;

-- 4. Reabilitar triggers
ALTER TABLE patients ENABLE TRIGGER enforce_patient_limit_before_insert;

-- 5. Verificar se ficou vazio
SELECT 'associations' as table_name, COUNT(*) as total FROM associations
UNION ALL SELECT 'users', COUNT(*) FROM users
UNION ALL SELECT 'patients', COUNT(*) FROM patients
UNION ALL SELECT 'doctors', COUNT(*) FROM doctors;

📝 Passo 2: Criar Associações de Teste
2.1 Criar Associação A (São Paulo)
bash

curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "association_name": "Associação Canábica de São Paulo",
    "cnpj": "12.345.678/0001-99",
    "phone": "(11) 99999-9999",
    "name": "João Silva",
    "email": "admin@spcannabis.com",
    "password": "123456"
  }' | jq '.'

Resposta esperada:
json

{
  "data": {
    "id": "...",
    "association_id": "f6f555b5-4deb-425c-acb3-9db9abbf8f23",
    "name": "João Silva",
    "email": "admin@spcannabis.com",
    "role": "admin",
    "is_active": true
  },
  "success": true
}

2.2 Criar Associação B (Rio de Janeiro)
bash

curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type": "application/json" \
  -d '{
    "association_name": "Associação Canábica do Rio de Janeiro",
    "cnpj": "98.765.432/0001-99",
    "phone": "(21) 88888-8888",
    "name": "Carlos Silva",
    "email": "admin@rjcannabis.com",
    "password": "123456"
  }' | jq '.'

2.3 Verificar no Banco
bash

docker exec -it cannacare_postgres psql -U postgres -d cannacare_db -c "
SELECT id, name, email FROM associations WHERE email LIKE '%cannabis%';
"

📝 Passo 3: Fazer Login e Obter Tokens
3.1 Login da Associação A
bash

curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@spcannabis.com","password":"123456"}' \
  | jq -r '.data.token' > tokenA.txt

3.2 Login da Associação B
bash

curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@rjcannabis.com","password":"123456"}' \
  | jq -r '.data.token' > tokenB.txt

3.3 Carregar Tokens
bash

TOKEN_A=$(cat tokenA.txt)
TOKEN_B=$(cat tokenB.txt)

echo "Token A: ${TOKEN_A:0:30}..."
echo "Token B: ${TOKEN_B:0:30}..."

📝 Passo 4: Criar Dados na Associação A
4.1 Criar Paciente A
bash

curl -X POST http://localhost:8080/api/patients \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "João Silva (Associação A)",
    "birth_date": "1980-01-01T00:00:00Z",
    "cpf": "87231084281",
    "phone": "(11) 99999-9999",
    "email": "joao@email.com"
  }' | jq '.'

4.2 Criar Médico A
bash

curl -X POST http://localhost:8080/api/doctors \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dra. Ana Paula (Associação A)",
    "crm": "12345",
    "crm_state": "SP",
    "specialty": "Neurologia"
  }' | jq '.'

📝 Passo 5: Criar Dados na Associação B
5.1 Criar Paciente B
bash

curl -X POST http://localhost:8080/api/patients \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Maria Souza (Associação B)",
    "birth_date": "1990-05-15T00:00:00Z",
    "cpf": "74203506557",
    "phone": "(21) 88888-8888",
    "email": "maria@email.com"
  }' | jq '.'

5.2 Criar Médico B
bash

curl -X POST http://localhost:8080/api/doctors \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dr. Roberto Santos (Associação B)",
    "crm": "67890",
    "crm_state": "RJ",
    "specialty": "Psiquiatria"
  }' | jq '.'

📝 Passo 6: TESTE DE ISOLAMENTO
6.1 Teste com a Associação A
bash

echo "=========================================="
echo "      TESTE DE ISOLAMENTO - ASSOCIAÇÃO A"
echo "=========================================="

echo ""
echo "📋 PACIENTES - Associação A (SP):"
curl -s -X GET http://localhost:8080/api/patients \
  -H "Authorization: Bearer $TOKEN_A" | jq '.data.items[].full_name'

echo ""
echo "👨‍⚕️ MÉDICOS - Associação A (SP):"
curl -s -X GET http://localhost:8080/api/doctors \
  -H "Authorization: Bearer $TOKEN_A" | jq '.data.items[].name'

Resultado esperado:
text

📋 PACIENTES - Associação A (SP):
"João Silva (Associação A)"

👨‍⚕️ MÉDICOS - Associação A (SP):
"Dra. Ana Paula (Associação A)"

6.2 Teste com a Associação B
bash

echo "=========================================="
echo "      TESTE DE ISOLAMENTO - ASSOCIAÇÃO B"
echo "=========================================="

echo ""
echo "📋 PACIENTES - Associação B (RJ):"
curl -s -X GET http://localhost:8080/api/patients \
  -H "Authorization: Bearer $TOKEN_B" | jq '.data.items[].full_name'

echo ""
echo "👨‍⚕️ MÉDICOS - Associação B (RJ):"
curl -s -X GET http://localhost:8080/api/doctors \
  -H "Authorization: Bearer $TOKEN_B" | jq '.data.items[].name'

Resultado esperado:
text

📋 PACIENTES - Associação B (RJ):
"Maria Souza (Associação B)"

👨‍⚕️ MÉDICOS - Associação B (RJ):
"Dr. Roberto Santos (Associação B)"

📝 Passo 7: Verificação no Banco de Dados
bash

docker exec -it cannacare_postgres psql -U postgres -d cannacare_db -c "
SELECT 
    a.name as associacao,
    COUNT(p.id) as pacientes,
    COUNT(d.id) as medicos
FROM associations a
LEFT JOIN patients p ON p.association_id = a.id
LEFT JOIN doctors d ON d.association_id = a.id
WHERE a.email LIKE '%cannabis%'
GROUP BY a.id, a.name
ORDER BY a.name;
"

Resultado esperado:
text

              associacao               | pacientes | medicos
---------------------------------------+-----------+---------
 Associação Canábica de São Paulo      |         1 |       1
 Associação Canábica do Rio de Janeiro |         1 |       1

📝 Passo 8: Resumo dos Testes
✅ Teste Final Consolidado
bash

echo "=========================================="
echo "      TESTE COMPLETO DE ISOLAMENTO"
echo "=========================================="

echo ""
echo "📋 PACIENTES A:"
curl -s -X GET http://localhost:8080/api/patients -H "Authorization: Bearer $TOKEN_A" | jq '.data.items[].full_name'

echo ""
echo "📋 PACIENTES B:"
curl -s -X GET http://localhost:8080/api/patients -H "Authorization: Bearer $TOKEN_B" | jq '.data.items[].full_name'

echo ""
echo "👨‍⚕️ MÉDICOS A:"
curl -s -X GET http://localhost:8080/api/doctors -H "Authorization: Bearer $TOKEN_A" | jq '.data.items[].name'

echo ""
echo "👨‍⚕️ MÉDICOS B:"
curl -s -X GET http://localhost:8080/api/doctors -H "Authorization: Bearer $TOKEN_B" | jq '.data.items[].name'

✅ Resultado Esperado
text

==========================================
      TESTE COMPLETO DE ISOLAMENTO
==========================================

📋 PACIENTES A:
"João Silva (Associação A)"

📋 PACIENTES B:
"Maria Souza (Associação B)"

👨‍⚕️ MÉDICOS A:
"Dra. Ana Paula (Associação A)"

👨‍⚕️ MÉDICOS B:
"Dr. Roberto Santos (Associação B)"

📊 Checklist de Validação
Item	Status	Descrição
✅	Criar Associação A	admin@spcannabis.com
✅	Criar Associação B	admin@rjcannabis.com
✅	Login A - Token gerado	Token salvo em tokenA.txt
✅	Login B - Token gerado	Token salvo em tokenB.txt
✅	Criar Paciente A	João Silva (Associação A)
✅	Criar Médico A	Dra. Ana Paula (Associação A)
✅	Criar Paciente B	Maria Souza (Associação B)
✅	Criar Médico B	Dr. Roberto Santos (Associação B)
✅	Listar Pacientes A	Só o paciente A
✅	Listar Pacientes B	Só o paciente B
✅	Listar Médicos A	Só o médico A
✅	Listar Médicos B	Só o médico B
🎯 Conclusão

O multi-tenancy está funcionando perfeitamente! Cada associação vê APENAS os seus próprios dados:

    Associação A vê apenas o paciente A e o médico A

    Associação B vê apenas o paciente B e o médico B

Isso confirma que o isolamento de dados está 100% funcional e seguro.
📝 Observações Finais
Principais Aprendizados

    O association_id é a chave do isolamento - Todas as tabelas têm esta coluna

    O JWT contém o association_id - O token sabe qual associação o usuário pertence

    O middleware extrai o association_id - Cada requisição tem o ID disponível

    Todas as queries filtram por association_id - NUNCA faça uma query sem este filtro

Comandos Úteis
bash

# Recriar usuário admin padrão
docker exec -it cannacare_postgres psql -U postgres -d cannacare_db -c "
INSERT INTO associations (id, name, cnpj, email, plan, status, patient_limit)
VALUES (gen_random_uuid(), 'CannaCare Admin', '00.000.000/0001-00', 'admin@cannacare.com', 'enterprise', 'active', 999999)
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (id, association_id, name, email, password_hash, role)
SELECT gen_random_uuid(), a.id, 'Administrador', 'admin@cannacare.com', 
'$2a$10$K8WrxL5rXB3p.OqM9zl5aeZ5EoW7JjLtYq2xNO6hVZ.SZ9y5Xb4bW', 'admin'
FROM associations a WHERE a.email = 'admin@cannacare.com'
ON CONFLICT (email, association_id) DO NOTHING;
"

🌿 CannaCare - Multi-Tenancy Testado e Aprovado!

```