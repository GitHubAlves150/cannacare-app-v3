-- ================================================================
-- CANNACARE - RESET DO BANCO (AMBIENTE DE TESTE)
-- ================================================================
-- ⚠️ ATENÇÃO: isso APAGA todos os dados de todas as tabelas.
-- Use só em ambiente de desenvolvimento/teste, nunca em produção.
--
-- O que faz:
--   1. Apaga todos os dados de todas as tabelas (TRUNCATE ... CASCADE)
--   2. Cria UMA associação base + o usuário admin
--      (adm@cannacare.com / admin123)
--
-- Depois de rodar, use esse admin pra logar no sistema e criar as
-- duas associações de teste (com CNPJ validado de verdade) pelo
-- fluxo normal do site, ou insira manualmente se preferir testar
-- só o backend.
-- ================================================================

BEGIN;

-- --- 1. Apagar todos os dados (CASCADE cuida da ordem das FKs) ---
TRUNCATE TABLE
    invite_tokens,
    patient_documents,
    patient_status_history,
    anamneses,
    notifications,
    order_items,
    orders,
    stock_movements,
    product_lots,
    prescription_items,
    prescriptions,
    products,
    doctors,
    payments,
    subscriptions,
    patients,
    users,
    associations
CASCADE;

-- --- 2. Criar associação base para o admin ---
INSERT INTO associations (
    id, name, cnpj, email, phone, address,
    plan, status, patient_limit, user_limit,
    plan_activated_at, plan_expires_at,
    created_at, updated_at
) VALUES (
    gen_random_uuid(),
    'CannaCare - Administração',
    '11111111000191',
    'adm@cannacare.com',
    '(00) 00000-0000',
    'Endereço interno',
    'enterprise',
    'active',
    999999,
    999999,
    now(),
    NULL, -- enterprise não expira
    now(),
    now()
);

-- --- 3. Criar o usuário admin ---
-- Senha: admin123
-- Hash gerado e verificado com bcrypt de verdade (custo 12) — não é
-- placeholder, foi testado batendo com "admin123" antes de ir pra cá.
INSERT INTO users (
    id, association_id, name, email, password_hash,
    role, is_active, created_at, updated_at
)
SELECT
    gen_random_uuid(),
    a.id,
    'Administrador',
    'adm@cannacare.com',
    '$2b$12$26hTf/TqgGSnht1vg/6mgeByPS.wozBQz3LyAwJsuxFA8/HtDKHvm',
    'admin',
    true,
    now(),
    now()
FROM associations a
WHERE a.cnpj = '11111111000191';

COMMIT;

-- --- Conferência ---
SELECT id, name, email, role, is_active FROM users;
SELECT id, name, cnpj, plan, status FROM associations;
