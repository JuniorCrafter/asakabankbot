-- Создание индексов для оптимизации запросов (Оптимизация производительности)

-- Индекс для быстрого поиска активных сессий клиента (используется в GetActiveSessionByClientTgID)
CREATE INDEX IF NOT EXISTS idx_chat_sessions_client_status ON public.chat_sessions (client_telegram_id, status);

-- Индекс для быстрого поиска активных сессий оператора (используется в GetActiveSessionByOperator)
CREATE INDEX IF NOT EXISTS idx_chat_sessions_operator_status ON public.chat_sessions (operator_id, status);

-- Индекс для быстрого поиска оператора по telegram_id
CREATE INDEX IF NOT EXISTS idx_operators_telegram_id ON public.operators (telegram_id);

-- Индекс для быстрого поиска пользователей по telegram_id (ускорит логику клиента)
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON public.users (telegram_id);

-- Индекс для поиска онлайн операторов по отделам (используется в RabbitMQ Consumer'e)
CREATE INDEX IF NOT EXISTS idx_operators_dep_status ON public.operators (department_id, status);
