CREATE TABLE recurring_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    category_id UUID REFERENCES categories(id) ON DELETE RESTRICT,
    mutation_type TEXT NOT NULL CHECK (mutation_type IN ('masuk', 'keluar')),
    amount NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    description TEXT NOT NULL,
    interval_unit TEXT NOT NULL CHECK (interval_unit IN ('daily', 'weekly', 'monthly')),
    interval_step INTEGER NOT NULL DEFAULT 1 CHECK (interval_step > 0),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ NOT NULL,
    last_run_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT recurring_rules_end_after_start CHECK (end_at IS NULL OR end_at >= start_at)
);

CREATE INDEX idx_recurring_rules_user_id_next_run_at
    ON recurring_rules (user_id, next_run_at)
    WHERE is_active = TRUE AND deleted_at IS NULL;

CREATE INDEX idx_recurring_rules_user_id_deleted_at
    ON recurring_rules (user_id, deleted_at);
