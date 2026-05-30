-- 1. СНАЧАЛА СОЗДАЕМ ВСЕ ТИПЫ (ENUMS)
CREATE TYPE public.bot_state_enum AS ENUM ('reg_menu', 'main_menu', 'in_dep', 'in_chat');
CREATE TYPE public.operator_status_enum AS ENUM ('online', 'offline', 'busy', 'admin_menu');
CREATE TYPE public.operator_stack_enum AS ENUM ('expert', 'expert_pro', 'expert_vip', 'expert_lite');
CREATE TYPE public.session_status_enum AS ENUM ('active', 'in_progress', 'closed');

-- 2. СОЗДАЕМ НЕЗАВИСИМЫЕ ТАБЛИЦЫ
CREATE TABLE public.users (
    id serial4 NOT NULL,
    telegram_id int8 NOT NULL,
    username varchar(255) NULL,
    "name" varchar(150) NULL,
    tel_number varchar(20) NULL,
    bot_state public."bot_state_enum" DEFAULT 'reg_menu'::bot_state_enum NULL,
    language_code varchar(10) NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_telegram_id_key UNIQUE (telegram_id)
);

CREATE TABLE public.departments (
    id serial4 NOT NULL,
    "name" varchar(100) NOT NULL,
    services jsonb DEFAULT '[]'::jsonb NOT NULL,
    CONSTRAINT departments_pkey PRIMARY KEY (id)
);

-- 3. СОЗДАЕМ ЗАВИСИМЫЕ ТАБЛИЦЫ
CREATE TABLE public.operators (
    id serial4 NOT NULL,
    telegram_id int8 NOT NULL,
    stack public."operator_stack_enum" NOT NULL,
    status public."operator_status_enum" DEFAULT 'offline'::operator_status_enum NULL,
    department_id int4 NULL,
    is_admin bool DEFAULT false NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
    CONSTRAINT operators_pkey PRIMARY KEY (id),
    CONSTRAINT operators_telegram_id_key UNIQUE (telegram_id)
);

CREATE TABLE public.chat_sessions (
    id serial4 NOT NULL,
    client_telegram_id int8 NULL,
    operator_id int4 NULL,
    department_id int4 NULL,
    service varchar(255) NOT NULL,
    status public."session_status_enum" DEFAULT 'active'::session_status_enum NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
    closed_at timestamptz NULL,
    chat_history jsonb DEFAULT '[]'::jsonb NOT NULL,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
    CONSTRAINT chat_sessions_pkey PRIMARY KEY (id)
);

-- 4. НАВЕШИВАЕМ ВНЕШНИЕ КЛЮЧИ (СВЯЗИ)
ALTER TABLE public.operators ADD CONSTRAINT operators_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.departments(id) ON DELETE SET NULL;
ALTER TABLE public.operators ADD CONSTRAINT operators_telegram_id_fkey FOREIGN KEY (telegram_id) REFERENCES public.users(telegram_id) ON DELETE CASCADE;
ALTER TABLE public.chat_sessions ADD CONSTRAINT chat_sessions_client_telegram_id_fkey FOREIGN KEY (client_telegram_id) REFERENCES public.users(telegram_id) ON DELETE CASCADE;
ALTER TABLE public.chat_sessions ADD CONSTRAINT chat_sessions_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.departments(id) ON DELETE SET NULL;
ALTER TABLE public.chat_sessions ADD CONSTRAINT chat_sessions_operator_id_fkey FOREIGN KEY (operator_id) REFERENCES public.operators(id) ON DELETE SET NULL;

INSERT INTO departments (name, services) VALUES 
('Физ. лица', '["Кредиты", "Вклады", "Карты", "Денежные переводы", "Курс валют", "Акции", "Запись онлайн", "Тарифы", "Asaka Travel", "ESG"]'::jsonb),
('Юр. лица', '["Кредиты", "Карты", "Депозиты", "Финансирование", "Эквайринг", "Фармацевтика", "Тарифы", "ESG", "Кредитные линии", "Интернет банкинг"]'::jsonb),
('Махалла банкирлари', '["Кредиты", "Вклады", "Карты", "Денежные переводы", "Курс валют", "Акции", "Запись онлайн", "Тарифы", "Asaka Travel", "ESG", "Депозиты", "Финансирование", "Эквайринг", "Фармацевтика", "Кредитные линии", "Интернет банкинг"]'::jsonb),
('Общие вопросы', '["Общий вопрос"]'::jsonb)
ON CONFLICT DO NOTHING;