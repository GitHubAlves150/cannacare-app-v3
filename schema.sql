--
-- PostgreSQL database dump
--

\restrict 7fY9evOMH4HSI0gdFlr7ACto4kC4a6s3jWSmwdKbKNAOes9VlLbUSjwy7E9qhpl

-- Dumped from database version 15.18
-- Dumped by pg_dump version 15.18

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: check_patient_limit(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.check_patient_limit(p_association_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    association_record associations%ROWTYPE;
    patient_count INTEGER;
BEGIN
    -- Buscar dados da associação
    SELECT * INTO association_record 
    FROM associations 
    WHERE id = p_association_id;
    
    -- Contar pacientes ativos da associação
    SELECT COUNT(*) INTO patient_count 
    FROM patients 
    WHERE association_id = p_association_id 
    AND deleted_at IS NULL;
    
    -- Verificar se atingiu o limite
    IF association_record.plan = 'basic' AND patient_count >= association_record.patient_limit THEN
        RAISE EXCEPTION 'Limite de pacientes do plano básico atingido. Faça upgrade para o plano premium.';
        RETURN FALSE;
    END IF;
    
    RETURN TRUE;
END;
$$;


--
-- Name: check_user_limit(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.check_user_limit(p_association_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    association_record associations%ROWTYPE;
    user_count INTEGER;
BEGIN
    SELECT * INTO association_record FROM associations WHERE id = p_association_id;

    SELECT COUNT(*) INTO user_count
    FROM users
    WHERE association_id = p_association_id
    AND is_active = true;

    IF user_count >= association_record.user_limit THEN
        RAISE EXCEPTION 'Limite de % usuários atingido para esta associação.', association_record.user_limit;
    END IF;

    RETURN TRUE;
END;
$$;


--
-- Name: enforce_patient_limit(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_patient_limit() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM check_patient_limit(NEW.association_id);
    RETURN NEW;
END;
$$;


--
-- Name: enforce_user_limit(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_user_limit() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.is_active = true THEN
        PERFORM check_user_limit(NEW.association_id);
    END IF;
    RETURN NEW;
END;
$$;


--
-- Name: process_order_stock(uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.process_order_stock(order_id uuid) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    item RECORD;
BEGIN
    -- Para cada item do pedido
    FOR item IN 
        SELECT oi.*, pl.current_quantity as current_stock
        FROM order_items oi
        JOIN product_lots pl ON pl.id = oi.product_lot_id
        WHERE oi.order_id = process_order_stock.order_id
    LOOP
        -- Verificar se há estoque suficiente
        IF item.current_stock < item.quantity THEN
            RAISE EXCEPTION 'Estoque insuficiente para o lote %', item.product_lot_id;
        END IF;
        
        -- Atualizar quantidade no lote
        UPDATE product_lots
        SET current_quantity = current_quantity - item.quantity
        WHERE id = item.product_lot_id;
        
        -- Registrar movimentação
        INSERT INTO stock_movements (
            product_lot_id,
            order_id,
            type,
            quantity,
            previous_quantity,
            new_quantity,
            user_id,
            notes,
            association_id
        ) VALUES (
            item.product_lot_id,
            process_order_stock.order_id,
            'baixa_pedido',
            -item.quantity,
            item.current_stock,
            item.current_stock - item.quantity,
            (SELECT user_id FROM orders WHERE id = process_order_stock.order_id),
            'Baixa automática do pedido',
            (SELECT association_id FROM orders WHERE id = process_order_stock.order_id)
        );
    END LOOP;
    
    -- Atualizar status do pedido para "separado"
    UPDATE orders
    SET status = 'separado',
        status_updated_at = CURRENT_TIMESTAMP
    WHERE id = process_order_stock.order_id;
END;
$$;


--
-- Name: set_default_user_limit(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_default_user_limit() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.user_limit IS NULL THEN
        NEW.user_limit := CASE 
            WHEN NEW.plan = 'premium' THEN 10
            ELSE 3
        END;
    END IF;
    RETURN NEW;
END;
$$;


--
-- Name: update_prescription_status(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_prescription_status() RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Atualizar para "proxima_vencer" (15 dias antes)
    UPDATE prescriptions
    SET status = 'proxima_vencer'
    WHERE expiration_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '15 days'
    AND status NOT IN ('proxima_vencer', 'vencida')
    AND is_active = true;
    
    -- Atualizar para "vencida"
    UPDATE prescriptions
    SET status = 'vencida',
        is_active = false
    WHERE expiration_date < CURRENT_DATE
    AND status != 'vencida';
END;
$$;


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: anamneses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.anamneses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    patient_id uuid NOT NULL,
    responsible_user_id uuid NOT NULL,
    type text NOT NULL,
    symptoms text,
    symptom_intensity bigint,
    side_effects text,
    side_effect_intensity bigint,
    treatment_adherence text,
    challenges text,
    improvements text,
    additional_notes text,
    weight numeric(5,2),
    blood_pressure text,
    heart_rate bigint,
    extra_responses jsonb,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT anamneses_side_effect_intensity_check CHECK (((side_effect_intensity >= 1) AND (side_effect_intensity <= 10))),
    CONSTRAINT anamneses_symptom_intensity_check CHECK (((symptom_intensity >= 1) AND (symptom_intensity <= 10))),
    CONSTRAINT anamneses_treatment_adherence_check CHECK ((treatment_adherence = ANY (ARRAY[('alta'::character varying)::text, ('media'::character varying)::text, ('baixa'::character varying)::text]))),
    CONSTRAINT anamneses_type_check CHECK ((type = ANY (ARRAY[('inicial'::character varying)::text, ('rastreio_1_mes'::character varying)::text, ('rastreio_3_meses'::character varying)::text, ('rastreio_6_meses'::character varying)::text, ('acompanhamento_continuo'::character varying)::text]))),
    CONSTRAINT chk_anamneses_side_effect_intensity CHECK (((side_effect_intensity >= 1) AND (side_effect_intensity <= 10))),
    CONSTRAINT chk_anamneses_symptom_intensity CHECK (((symptom_intensity >= 1) AND (symptom_intensity <= 10))),
    CONSTRAINT chk_anamneses_treatment_adherence CHECK ((treatment_adherence = ANY (ARRAY['alta'::text, 'media'::text, 'baixa'::text])))
);


--
-- Name: associations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.associations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    cnpj text NOT NULL,
    email text NOT NULL,
    phone text,
    address text,
    plan text DEFAULT 'basic'::character varying,
    status text DEFAULT 'pending'::character varying,
    patient_limit bigint DEFAULT 50,
    trial_ends_at timestamp without time zone,
    stripe_customer_id character varying(100),
    subscription_id character varying(100),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    user_limit bigint DEFAULT 3,
    plan_activated_at timestamp without time zone,
    plan_expires_at timestamp without time zone,
    payment_reference text,
    CONSTRAINT associations_plan_check CHECK ((plan = ANY (ARRAY[('basic'::character varying)::text, ('premium'::character varying)::text, ('enterprise'::character varying)::text]))),
    CONSTRAINT associations_status_check CHECK ((status = ANY (ARRAY[('pending'::character varying)::text, ('active'::character varying)::text, ('suspended'::character varying)::text, ('cancelled'::character varying)::text]))),
    CONSTRAINT chk_associations_plan CHECK ((plan = ANY (ARRAY['basic'::text, 'premium'::text, 'enterprise'::text]))),
    CONSTRAINT chk_associations_status CHECK ((status = ANY (ARRAY['pending'::text, 'active'::text, 'suspended'::text, 'cancelled'::text])))
);


--
-- Name: doctors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.doctors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    name text NOT NULL,
    crm text NOT NULL,
    crm_state text NOT NULL,
    specialty text,
    phone text,
    email text,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone
);


--
-- Name: invite_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invite_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash character varying(64) NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    used_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now()
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    user_id uuid,
    patient_id uuid,
    type text NOT NULL,
    title text NOT NULL,
    message text NOT NULL,
    read_at timestamp without time zone,
    action_url text,
    created_at timestamp without time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT notifications_type_check CHECK ((type = ANY (ARRAY[('prescription_expiring'::character varying)::text, ('prescription_expired'::character varying)::text, ('low_stock'::character varying)::text, ('product_expiring'::character varying)::text, ('payment_due'::character varying)::text, ('order_status'::character varying)::text])))
);


--
-- Name: order_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    order_id uuid NOT NULL,
    product_lot_id uuid NOT NULL,
    quantity integer NOT NULL,
    unit_price numeric(10,2) NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    patient_id uuid NOT NULL,
    prescription_id uuid NOT NULL,
    status text DEFAULT 'pendente'::character varying NOT NULL,
    shipping_carrier text,
    tracking_code text,
    shipping_label_url text,
    shipping_cost numeric(10,2) DEFAULT 0,
    total_amount numeric(10,2) NOT NULL,
    notes text,
    order_date timestamp without time zone,
    status_updated_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    CONSTRAINT orders_status_check CHECK ((status = ANY (ARRAY[('pendente'::character varying)::text, ('separado'::character varying)::text, ('dispensa'::character varying)::text, ('correio'::character varying)::text, ('entregue'::character varying)::text, ('cancelado'::character varying)::text])))
);


--
-- Name: patient_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.patient_documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    patient_id uuid NOT NULL,
    document_type text NOT NULL,
    file_url text NOT NULL,
    file_name text,
    file_size bigint,
    mime_type text,
    status text DEFAULT 'em_analise'::character varying,
    reviewed_by uuid,
    reviewed_at text,
    created_at timestamp without time zone,
    deleted_at timestamp without time zone,
    updated_at timestamp with time zone,
    CONSTRAINT patient_documents_document_type_check CHECK ((document_type = ANY (ARRAY[('rg_cpf'::character varying)::text, ('comprovante_residencia'::character varying)::text, ('laudo_medico'::character varying)::text, ('receita_medica'::character varying)::text, ('autorizacao_anvisa'::character varying)::text, ('termo_consentimento'::character varying)::text]))),
    CONSTRAINT patient_documents_status_check CHECK ((status = ANY (ARRAY[('em_analise'::character varying)::text, ('aprovado'::character varying)::text, ('rejeitado'::character varying)::text])))
);


--
-- Name: patient_status_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.patient_status_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    patient_id uuid NOT NULL,
    old_status text,
    new_status text NOT NULL,
    changed_by uuid,
    reason text,
    created_at timestamp without time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: patients; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.patients (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    user_id uuid,
    full_name text NOT NULL,
    birth_date timestamp with time zone NOT NULL,
    gender text,
    cpf text NOT NULL,
    rg text,
    phone text,
    whatsapp text,
    email text,
    address_street text,
    address_number text,
    address_complement text,
    address_neighborhood text,
    address_city text,
    address_state text,
    address_zipcode character varying(10),
    status text DEFAULT 'pendente_documentacao'::character varying NOT NULL,
    is_social_patient boolean DEFAULT false,
    social_assistant_notes text,
    approved_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    address_zip_code text,
    CONSTRAINT patients_status_check CHECK ((status = ANY (ARRAY[('pendente_documentacao'::character varying)::text, ('em_analise'::character varying)::text, ('aprovado'::character varying)::text, ('negado'::character varying)::text, ('assistente_social'::character varying)::text])))
);


--
-- Name: payments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    patient_id uuid NOT NULL,
    order_id uuid,
    subscription_id uuid,
    payment_type character varying(20) NOT NULL,
    payment_method character varying(20) NOT NULL,
    amount numeric(10,2) NOT NULL,
    installments integer DEFAULT 1,
    receipt_url text,
    receipt_number character varying(50),
    status character varying(20) DEFAULT 'pendente'::character varying,
    payment_date date,
    paid_at timestamp without time zone,
    transaction_id character varying(100),
    gateway_response jsonb,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone,
    CONSTRAINT payment_source_check CHECK ((((order_id IS NOT NULL) AND (subscription_id IS NULL)) OR ((order_id IS NULL) AND (subscription_id IS NOT NULL)) OR (((payment_type)::text = 'doacao'::text) AND (order_id IS NULL) AND (subscription_id IS NULL)))),
    CONSTRAINT payments_payment_method_check CHECK (((payment_method)::text = ANY ((ARRAY['pix'::character varying, 'boleto'::character varying, 'cartao'::character varying, 'transferencia'::character varying])::text[]))),
    CONSTRAINT payments_payment_type_check CHECK (((payment_type)::text = ANY ((ARRAY['anuidade'::character varying, 'compra_produto'::character varying, 'doacao'::character varying])::text[]))),
    CONSTRAINT payments_status_check CHECK (((status)::text = ANY ((ARRAY['pendente'::character varying, 'pago'::character varying, 'recusado'::character varying, 'estornado'::character varying])::text[])))
);


--
-- Name: prescription_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prescription_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    prescription_id uuid NOT NULL,
    product_id uuid NOT NULL,
    dosage_instructions text NOT NULL,
    quantity_recommended bigint NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: prescriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prescriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    patient_id uuid NOT NULL,
    doctor_id uuid NOT NULL,
    cid text NOT NULL,
    issue_date timestamp with time zone NOT NULL,
    expiration_date timestamp with time zone NOT NULL,
    status text DEFAULT 'valida'::character varying,
    is_active boolean DEFAULT true,
    prescription_file_url text,
    prescription_file_name text,
    validated_by uuid,
    validated_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    CONSTRAINT prescriptions_status_check CHECK ((status = ANY (ARRAY[('valida'::character varying)::text, ('proxima_vencer'::character varying)::text, ('vencida'::character varying)::text]))),
    CONSTRAINT valid_expiration CHECK ((expiration_date >= issue_date))
);


--
-- Name: product_lots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.product_lots (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    product_id uuid NOT NULL,
    lot_number text NOT NULL,
    expiration_date timestamp with time zone NOT NULL,
    current_quantity bigint DEFAULT 0 NOT NULL,
    initial_quantity bigint DEFAULT 0 NOT NULL,
    supplier text,
    purchase_date timestamp with time zone,
    purchase_price numeric(10,2),
    received_by uuid,
    received_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone
);


--
-- Name: products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    unit_price numeric(10,2) NOT NULL,
    min_stock_alert bigint DEFAULT 10,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone
);


--
-- Name: stock_movements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stock_movements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    product_lot_id uuid NOT NULL,
    order_id uuid,
    type text NOT NULL,
    quantity bigint NOT NULL,
    previous_quantity bigint NOT NULL,
    new_quantity bigint NOT NULL,
    user_id uuid NOT NULL,
    notes text,
    created_at timestamp without time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT stock_movements_type_check CHECK ((type = ANY (ARRAY[('entrada'::character varying)::text, ('baixa_pedido'::character varying)::text, ('ajuste_manual'::character varying)::text, ('perda'::character varying)::text, ('devolucao'::character varying)::text])))
);


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    patient_id uuid NOT NULL,
    due_date date NOT NULL,
    amount numeric(10,2) NOT NULL,
    status character varying(20) DEFAULT 'pendente'::character varying,
    paid_at timestamp without time zone,
    payment_id uuid,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone,
    CONSTRAINT subscriptions_status_check CHECK (((status)::text = ANY ((ARRAY['pendente'::character varying, 'pago'::character varying, 'atrasado'::character varying, 'cancelado'::character varying])::text[])))
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    association_id uuid NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    role text NOT NULL,
    is_active boolean DEFAULT true,
    last_login_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    CONSTRAINT users_role_check CHECK ((role = ANY (ARRAY[('admin'::character varying)::text, ('secretaria'::character varying)::text, ('acolhimento'::character varying)::text, ('farmacia'::character varying)::text, ('coordenacao'::character varying)::text])))
);


--
-- Name: vw_expired_prescriptions; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.vw_expired_prescriptions AS
 SELECT p.id AS patient_id,
    p.full_name,
    p.cpf,
    p.phone,
    pr.id AS prescription_id,
    pr.expiration_date,
    pr.doctor_id,
    d.name AS doctor_name,
    d.crm,
    p.association_id,
    EXTRACT(day FROM age((CURRENT_DATE)::timestamp with time zone, pr.expiration_date)) AS days_expired
   FROM ((public.patients p
     JOIN public.prescriptions pr ON ((pr.patient_id = p.id)))
     JOIN public.doctors d ON ((d.id = pr.doctor_id)))
  WHERE ((pr.expiration_date < CURRENT_DATE) AND (pr.is_active = true) AND (p.status = 'aprovado'::text) AND (p.deleted_at IS NULL))
  ORDER BY pr.expiration_date;


--
-- Name: vw_low_stock; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.vw_low_stock AS
 SELECT pl.id AS lot_id,
    p.id AS product_id,
    p.name AS product_name,
    pl.lot_number,
    pl.expiration_date,
    pl.current_quantity,
    p.min_stock_alert,
    (p.min_stock_alert - pl.current_quantity) AS missing_units,
    p.association_id
   FROM (public.product_lots pl
     JOIN public.products p ON ((p.id = pl.product_id)))
  WHERE ((pl.current_quantity <= p.min_stock_alert) AND (pl.expiration_date > CURRENT_DATE) AND (pl.deleted_at IS NULL))
  ORDER BY (p.min_stock_alert - pl.current_quantity) DESC;


--
-- Name: vw_overdue_subscriptions; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.vw_overdue_subscriptions AS
 SELECT s.id AS subscription_id,
    p.id AS patient_id,
    p.full_name,
    p.phone,
    s.due_date,
    s.amount,
    p.association_id,
    EXTRACT(day FROM age((CURRENT_DATE)::timestamp with time zone, (s.due_date)::timestamp with time zone)) AS days_overdue
   FROM (public.subscriptions s
     JOIN public.patients p ON ((p.id = s.patient_id)))
  WHERE (((s.status)::text = 'atrasado'::text) AND (s.deleted_at IS NULL))
  ORDER BY s.due_date;


--
-- Name: vw_patient_dashboard; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.vw_patient_dashboard AS
 SELECT patients.association_id,
    patients.status,
    count(*) AS total,
    count(
        CASE
            WHEN (patients.is_social_patient = true) THEN 1
            ELSE NULL::integer
        END) AS social_patients,
    count(
        CASE
            WHEN (patients.created_at >= (CURRENT_DATE - '30 days'::interval)) THEN 1
            ELSE NULL::integer
        END) AS last_30_days
   FROM public.patients
  WHERE (patients.deleted_at IS NULL)
  GROUP BY patients.association_id, patients.status;


--
-- Name: vw_stock_summary; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.vw_stock_summary AS
 SELECT p.id AS product_id,
    p.name AS product_name,
    p.association_id,
    count(DISTINCT pl.id) AS total_lots,
    COALESCE(sum(pl.current_quantity), (0)::numeric) AS total_quantity,
    COALESCE(avg(pl.current_quantity), (0)::numeric) AS avg_per_lot,
    min(pl.expiration_date) AS earliest_expiration
   FROM (public.products p
     LEFT JOIN public.product_lots pl ON ((pl.product_id = p.id)))
  WHERE ((p.is_active = true) AND (p.deleted_at IS NULL))
  GROUP BY p.id, p.name, p.association_id
  ORDER BY p.name;


--
-- Name: vw_top_doctors; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.vw_top_doctors AS
 SELECT d.id AS doctor_id,
    d.name AS doctor_name,
    d.crm,
    d.specialty,
    d.association_id,
    count(pr.id) AS total_prescriptions,
    count(DISTINCT pr.patient_id) AS unique_patients
   FROM (public.doctors d
     LEFT JOIN public.prescriptions pr ON ((pr.doctor_id = d.id)))
  WHERE ((d.is_active = true) AND (d.deleted_at IS NULL))
  GROUP BY d.id, d.name, d.crm, d.specialty, d.association_id
  ORDER BY (count(pr.id)) DESC;


--
-- Name: anamneses anamneses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.anamneses
    ADD CONSTRAINT anamneses_pkey PRIMARY KEY (id);


--
-- Name: associations associations_cnpj_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.associations
    ADD CONSTRAINT associations_cnpj_key UNIQUE (cnpj);


--
-- Name: associations associations_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.associations
    ADD CONSTRAINT associations_email_key UNIQUE (email);


--
-- Name: associations associations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.associations
    ADD CONSTRAINT associations_pkey PRIMARY KEY (id);


--
-- Name: doctors doctors_crm_crm_state_association_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doctors
    ADD CONSTRAINT doctors_crm_crm_state_association_id_key UNIQUE (crm, crm_state, association_id);


--
-- Name: doctors doctors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doctors
    ADD CONSTRAINT doctors_pkey PRIMARY KEY (id);


--
-- Name: invite_tokens invite_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invite_tokens
    ADD CONSTRAINT invite_tokens_pkey PRIMARY KEY (id);


--
-- Name: invite_tokens invite_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invite_tokens
    ADD CONSTRAINT invite_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: patient_documents patient_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_documents
    ADD CONSTRAINT patient_documents_pkey PRIMARY KEY (id);


--
-- Name: patient_status_history patient_status_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_status_history
    ADD CONSTRAINT patient_status_history_pkey PRIMARY KEY (id);


--
-- Name: patients patients_cpf_association_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_cpf_association_id_key UNIQUE (cpf, association_id);


--
-- Name: patients patients_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_pkey PRIMARY KEY (id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: prescription_items prescription_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescription_items
    ADD CONSTRAINT prescription_items_pkey PRIMARY KEY (id);


--
-- Name: prescription_items prescription_items_prescription_id_product_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescription_items
    ADD CONSTRAINT prescription_items_prescription_id_product_id_key UNIQUE (prescription_id, product_id);


--
-- Name: prescriptions prescriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescriptions
    ADD CONSTRAINT prescriptions_pkey PRIMARY KEY (id);


--
-- Name: product_lots product_lots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_lots
    ADD CONSTRAINT product_lots_pkey PRIMARY KEY (id);


--
-- Name: product_lots product_lots_product_id_lot_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_lots
    ADD CONSTRAINT product_lots_product_id_lot_number_key UNIQUE (product_id, lot_number);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: stock_movements stock_movements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT stock_movements_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: patients uni_patients_user_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT uni_patients_user_id UNIQUE (user_id);


--
-- Name: users users_email_association_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_association_id_key UNIQUE (email, association_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_anamneses_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_anamneses_association ON public.anamneses USING btree (association_id);


--
-- Name: idx_anamneses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_anamneses_deleted_at ON public.anamneses USING btree (deleted_at);


--
-- Name: idx_anamneses_patient; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_anamneses_patient ON public.anamneses USING btree (patient_id);


--
-- Name: idx_anamneses_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_anamneses_type ON public.anamneses USING btree (type);


--
-- Name: idx_associations_cnpj; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_associations_cnpj ON public.associations USING btree (cnpj);


--
-- Name: idx_associations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_associations_deleted_at ON public.associations USING btree (deleted_at);


--
-- Name: idx_associations_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_associations_email ON public.associations USING btree (email);


--
-- Name: idx_associations_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_associations_status ON public.associations USING btree (status);


--
-- Name: idx_doctors_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doctors_association ON public.doctors USING btree (association_id);


--
-- Name: idx_doctors_association_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doctors_association_id ON public.doctors USING btree (association_id);


--
-- Name: idx_doctors_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doctors_deleted_at ON public.doctors USING btree (deleted_at);


--
-- Name: idx_invite_tokens_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_invite_tokens_user_id ON public.invite_tokens USING btree (user_id);


--
-- Name: idx_notifications_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_association ON public.notifications USING btree (association_id);


--
-- Name: idx_notifications_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_deleted_at ON public.notifications USING btree (deleted_at);


--
-- Name: idx_notifications_patient; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_patient ON public.notifications USING btree (patient_id);


--
-- Name: idx_notifications_read; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_read ON public.notifications USING btree (read_at);


--
-- Name: idx_notifications_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user ON public.notifications USING btree (user_id);


--
-- Name: idx_order_items_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_order_items_association ON public.order_items USING btree (association_id);


--
-- Name: idx_order_items_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_order_items_order ON public.order_items USING btree (order_id);


--
-- Name: idx_orders_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_association ON public.orders USING btree (association_id);


--
-- Name: idx_orders_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_deleted_at ON public.orders USING btree (deleted_at);


--
-- Name: idx_orders_patient; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_patient ON public.orders USING btree (patient_id);


--
-- Name: idx_orders_prescription; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_prescription ON public.orders USING btree (prescription_id);


--
-- Name: idx_orders_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_status ON public.orders USING btree (status);


--
-- Name: idx_patient_documents_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patient_documents_association ON public.patient_documents USING btree (association_id);


--
-- Name: idx_patient_documents_association_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patient_documents_association_id ON public.patient_documents USING btree (association_id);


--
-- Name: idx_patient_documents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patient_documents_deleted_at ON public.patient_documents USING btree (deleted_at);


--
-- Name: idx_patient_documents_patient; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patient_documents_patient ON public.patient_documents USING btree (patient_id);


--
-- Name: idx_patient_status_history_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patient_status_history_association ON public.patient_status_history USING btree (association_id);


--
-- Name: idx_patient_status_history_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patient_status_history_deleted_at ON public.patient_status_history USING btree (deleted_at);


--
-- Name: idx_patient_status_history_patient; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patient_status_history_patient ON public.patient_status_history USING btree (patient_id);


--
-- Name: idx_patients_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patients_association ON public.patients USING btree (association_id);


--
-- Name: idx_patients_association_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patients_association_id ON public.patients USING btree (association_id);


--
-- Name: idx_patients_cpf; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patients_cpf ON public.patients USING btree (cpf);


--
-- Name: idx_patients_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patients_deleted_at ON public.patients USING btree (deleted_at);


--
-- Name: idx_patients_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_patients_status ON public.patients USING btree (status);


--
-- Name: idx_payments_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payments_association ON public.payments USING btree (association_id);


--
-- Name: idx_payments_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payments_order ON public.payments USING btree (order_id);


--
-- Name: idx_payments_patient; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payments_patient ON public.payments USING btree (patient_id);


--
-- Name: idx_payments_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payments_status ON public.payments USING btree (status);


--
-- Name: idx_prescription_items_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prescription_items_association ON public.prescription_items USING btree (association_id);


--
-- Name: idx_prescription_items_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prescription_items_deleted_at ON public.prescription_items USING btree (deleted_at);


--
-- Name: idx_prescription_items_prescription; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prescription_items_prescription ON public.prescription_items USING btree (prescription_id);


--
-- Name: idx_prescriptions_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prescriptions_association ON public.prescriptions USING btree (association_id);


--
-- Name: idx_prescriptions_association_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prescriptions_association_id ON public.prescriptions USING btree (association_id);


--
-- Name: idx_prescriptions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prescriptions_deleted_at ON public.prescriptions USING btree (deleted_at);


--
-- Name: idx_prescriptions_expiration; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prescriptions_expiration ON public.prescriptions USING btree (expiration_date);


--
-- Name: idx_prescriptions_patient; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prescriptions_patient ON public.prescriptions USING btree (patient_id);


--
-- Name: idx_prescriptions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_prescriptions_status ON public.prescriptions USING btree (status);


--
-- Name: idx_product_lots_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_product_lots_association ON public.product_lots USING btree (association_id);


--
-- Name: idx_product_lots_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_product_lots_deleted_at ON public.product_lots USING btree (deleted_at);


--
-- Name: idx_product_lots_expiration; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_product_lots_expiration ON public.product_lots USING btree (expiration_date);


--
-- Name: idx_product_lots_product; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_product_lots_product ON public.product_lots USING btree (product_id);


--
-- Name: idx_products_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_association ON public.products USING btree (association_id);


--
-- Name: idx_products_association_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_association_id ON public.products USING btree (association_id);


--
-- Name: idx_products_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_deleted_at ON public.products USING btree (deleted_at);


--
-- Name: idx_stock_movements_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_movements_association ON public.stock_movements USING btree (association_id);


--
-- Name: idx_stock_movements_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_movements_deleted_at ON public.stock_movements USING btree (deleted_at);


--
-- Name: idx_stock_movements_lot; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_movements_lot ON public.stock_movements USING btree (product_lot_id);


--
-- Name: idx_stock_movements_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_movements_order ON public.stock_movements USING btree (order_id);


--
-- Name: idx_subscriptions_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_subscriptions_association ON public.subscriptions USING btree (association_id);


--
-- Name: idx_subscriptions_patient; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_subscriptions_patient ON public.subscriptions USING btree (patient_id);


--
-- Name: idx_subscriptions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_subscriptions_status ON public.subscriptions USING btree (status);


--
-- Name: idx_users_association; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_association ON public.users USING btree (association_id);


--
-- Name: idx_users_association_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_association_id ON public.users USING btree (association_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: patients enforce_patient_limit_before_insert; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER enforce_patient_limit_before_insert BEFORE INSERT ON public.patients FOR EACH ROW EXECUTE FUNCTION public.enforce_patient_limit();


--
-- Name: users enforce_user_limit_before_insert; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER enforce_user_limit_before_insert BEFORE INSERT ON public.users FOR EACH ROW EXECUTE FUNCTION public.enforce_user_limit();


--
-- Name: associations set_user_limit_before_insert; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER set_user_limit_before_insert BEFORE INSERT ON public.associations FOR EACH ROW EXECUTE FUNCTION public.set_default_user_limit();


--
-- Name: anamneses update_anamneses_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_anamneses_updated_at BEFORE UPDATE ON public.anamneses FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: associations update_associations_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_associations_updated_at BEFORE UPDATE ON public.associations FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: doctors update_doctors_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_doctors_updated_at BEFORE UPDATE ON public.doctors FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: orders update_orders_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON public.orders FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: patients update_patients_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_patients_updated_at BEFORE UPDATE ON public.patients FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: payments update_payments_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON public.payments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: prescriptions update_prescriptions_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_prescriptions_updated_at BEFORE UPDATE ON public.prescriptions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: product_lots update_product_lots_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_product_lots_updated_at BEFORE UPDATE ON public.product_lots FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: products update_products_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON public.products FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: subscriptions update_subscriptions_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON public.subscriptions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: anamneses anamneses_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.anamneses
    ADD CONSTRAINT anamneses_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: anamneses anamneses_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.anamneses
    ADD CONSTRAINT anamneses_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- Name: anamneses anamneses_responsible_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.anamneses
    ADD CONSTRAINT anamneses_responsible_user_id_fkey FOREIGN KEY (responsible_user_id) REFERENCES public.users(id);


--
-- Name: doctors doctors_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doctors
    ADD CONSTRAINT doctors_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: anamneses fk_anamneses_responsible_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.anamneses
    ADD CONSTRAINT fk_anamneses_responsible_user FOREIGN KEY (responsible_user_id) REFERENCES public.users(id);


--
-- Name: doctors fk_doctors_association; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doctors
    ADD CONSTRAINT fk_doctors_association FOREIGN KEY (association_id) REFERENCES public.associations(id);


--
-- Name: prescriptions fk_doctors_prescriptions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescriptions
    ADD CONSTRAINT fk_doctors_prescriptions FOREIGN KEY (doctor_id) REFERENCES public.doctors(id);


--
-- Name: notifications fk_notifications_patient; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_patient FOREIGN KEY (patient_id) REFERENCES public.patients(id);


--
-- Name: notifications fk_notifications_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: patient_documents fk_patient_documents_reviewer; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_documents
    ADD CONSTRAINT fk_patient_documents_reviewer FOREIGN KEY (reviewed_by) REFERENCES public.users(id);


--
-- Name: patient_status_history fk_patient_status_history_changed_by_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_status_history
    ADD CONSTRAINT fk_patient_status_history_changed_by_user FOREIGN KEY (changed_by) REFERENCES public.users(id);


--
-- Name: patient_status_history fk_patient_status_history_patient; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_status_history
    ADD CONSTRAINT fk_patient_status_history_patient FOREIGN KEY (patient_id) REFERENCES public.patients(id);


--
-- Name: anamneses fk_patients_anamneses; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.anamneses
    ADD CONSTRAINT fk_patients_anamneses FOREIGN KEY (patient_id) REFERENCES public.patients(id);


--
-- Name: patients fk_patients_association; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT fk_patients_association FOREIGN KEY (association_id) REFERENCES public.associations(id);


--
-- Name: patient_documents fk_patients_documents; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_documents
    ADD CONSTRAINT fk_patients_documents FOREIGN KEY (patient_id) REFERENCES public.patients(id);


--
-- Name: orders fk_patients_orders; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_patients_orders FOREIGN KEY (patient_id) REFERENCES public.patients(id);


--
-- Name: prescriptions fk_patients_prescriptions; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescriptions
    ADD CONSTRAINT fk_patients_prescriptions FOREIGN KEY (patient_id) REFERENCES public.patients(id);


--
-- Name: prescriptions fk_prescriptions_association; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescriptions
    ADD CONSTRAINT fk_prescriptions_association FOREIGN KEY (association_id) REFERENCES public.associations(id);


--
-- Name: prescription_items fk_prescriptions_items; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescription_items
    ADD CONSTRAINT fk_prescriptions_items FOREIGN KEY (prescription_id) REFERENCES public.prescriptions(id);


--
-- Name: orders fk_prescriptions_orders; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_prescriptions_orders FOREIGN KEY (prescription_id) REFERENCES public.prescriptions(id);


--
-- Name: stock_movements fk_product_lots_stock_movements; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT fk_product_lots_stock_movements FOREIGN KEY (product_lot_id) REFERENCES public.product_lots(id);


--
-- Name: products fk_products_association; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_products_association FOREIGN KEY (association_id) REFERENCES public.associations(id);


--
-- Name: product_lots fk_products_lots; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_lots
    ADD CONSTRAINT fk_products_lots FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: prescription_items fk_products_prescription_items; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescription_items
    ADD CONSTRAINT fk_products_prescription_items FOREIGN KEY (product_id) REFERENCES public.products(id);


--
-- Name: stock_movements fk_stock_movements_order; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT fk_stock_movements_order FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: stock_movements fk_stock_movements_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT fk_stock_movements_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: users fk_users_association; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_users_association FOREIGN KEY (association_id) REFERENCES public.associations(id);


--
-- Name: patients fk_users_patient; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT fk_users_patient FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: invite_tokens invite_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invite_tokens
    ADD CONSTRAINT invite_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: order_items order_items_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: order_items order_items_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: order_items order_items_product_lot_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_product_lot_id_fkey FOREIGN KEY (product_lot_id) REFERENCES public.product_lots(id);


--
-- Name: orders orders_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: orders orders_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- Name: orders orders_prescription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_prescription_id_fkey FOREIGN KEY (prescription_id) REFERENCES public.prescriptions(id);


--
-- Name: patient_documents patient_documents_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_documents
    ADD CONSTRAINT patient_documents_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: patient_documents patient_documents_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_documents
    ADD CONSTRAINT patient_documents_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- Name: patient_documents patient_documents_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_documents
    ADD CONSTRAINT patient_documents_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id);


--
-- Name: patient_status_history patient_status_history_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_status_history
    ADD CONSTRAINT patient_status_history_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: patient_status_history patient_status_history_changed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_status_history
    ADD CONSTRAINT patient_status_history_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES public.users(id);


--
-- Name: patient_status_history patient_status_history_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patient_status_history
    ADD CONSTRAINT patient_status_history_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- Name: patients patients_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: patients patients_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: payments payments_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: payments payments_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE SET NULL;


--
-- Name: payments payments_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- Name: payments payments_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_subscription_id_fkey FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(id) ON DELETE SET NULL;


--
-- Name: prescription_items prescription_items_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescription_items
    ADD CONSTRAINT prescription_items_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: prescription_items prescription_items_prescription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescription_items
    ADD CONSTRAINT prescription_items_prescription_id_fkey FOREIGN KEY (prescription_id) REFERENCES public.prescriptions(id) ON DELETE CASCADE;


--
-- Name: prescriptions prescriptions_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescriptions
    ADD CONSTRAINT prescriptions_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: prescriptions prescriptions_doctor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescriptions
    ADD CONSTRAINT prescriptions_doctor_id_fkey FOREIGN KEY (doctor_id) REFERENCES public.doctors(id);


--
-- Name: prescriptions prescriptions_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescriptions
    ADD CONSTRAINT prescriptions_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- Name: prescriptions prescriptions_validated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prescriptions
    ADD CONSTRAINT prescriptions_validated_by_fkey FOREIGN KEY (validated_by) REFERENCES public.users(id);


--
-- Name: product_lots product_lots_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_lots
    ADD CONSTRAINT product_lots_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: product_lots product_lots_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_lots
    ADD CONSTRAINT product_lots_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: product_lots product_lots_received_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_lots
    ADD CONSTRAINT product_lots_received_by_fkey FOREIGN KEY (received_by) REFERENCES public.users(id);


--
-- Name: products products_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: stock_movements stock_movements_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT stock_movements_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: stock_movements stock_movements_product_lot_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT stock_movements_product_lot_id_fkey FOREIGN KEY (product_lot_id) REFERENCES public.product_lots(id) ON DELETE CASCADE;


--
-- Name: stock_movements stock_movements_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT stock_movements_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: subscriptions subscriptions_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- Name: subscriptions subscriptions_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE CASCADE;


--
-- Name: users users_association_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_association_id_fkey FOREIGN KEY (association_id) REFERENCES public.associations(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict 7fY9evOMH4HSI0gdFlr7ACto4kC4a6s3jWSmwdKbKNAOes9VlLbUSjwy7E9qhpl

