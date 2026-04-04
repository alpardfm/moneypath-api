ALTER TABLE wallets
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE debts
    ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_wallets_user_id_deleted_at
    ON wallets (user_id, deleted_at);

CREATE INDEX idx_debts_user_id_deleted_at
    ON debts (user_id, deleted_at);
