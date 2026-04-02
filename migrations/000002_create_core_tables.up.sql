CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    balance NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT wallets_user_id_name_key UNIQUE (user_id, name)
);

CREATE INDEX idx_wallets_user_id_is_active
    ON wallets (user_id, is_active);

CREATE TABLE debts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    principal_amount NUMERIC(18,2) NOT NULL CHECK (principal_amount > 0),
    remaining_amount NUMERIC(18,2) NOT NULL CHECK (remaining_amount >= 0),
    tenor_value INTEGER,
    tenor_unit TEXT CHECK (tenor_unit IN ('day', 'week', 'month', 'year')),
    payment_amount NUMERIC(18,2) CHECK (payment_amount >= 0),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'lunas', 'inactive')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_debts_user_id_status
    ON debts (user_id, status);

CREATE INDEX idx_debts_user_id_is_active
    ON debts (user_id, is_active);

CREATE TABLE mutations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    debt_id UUID REFERENCES debts(id) ON DELETE RESTRICT,
    mutation_type TEXT NOT NULL CHECK (mutation_type IN ('masuk', 'keluar')),
    amount NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    description TEXT NOT NULL,
    related_to_debt BOOLEAN NOT NULL DEFAULT FALSE,
    happened_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT mutations_debt_relation_check CHECK (
        (related_to_debt = FALSE AND debt_id IS NULL) OR
        (related_to_debt = TRUE AND debt_id IS NOT NULL)
    )
);

CREATE INDEX idx_mutations_user_id_happened_at
    ON mutations (user_id, happened_at DESC);

CREATE INDEX idx_mutations_wallet_id_happened_at
    ON mutations (wallet_id, happened_at DESC);

CREATE INDEX idx_mutations_debt_id_happened_at
    ON mutations (debt_id, happened_at DESC);
