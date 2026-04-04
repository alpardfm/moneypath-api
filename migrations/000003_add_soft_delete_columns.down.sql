DROP INDEX IF EXISTS idx_debts_user_id_deleted_at;
DROP INDEX IF EXISTS idx_wallets_user_id_deleted_at;

ALTER TABLE debts
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE wallets
    DROP COLUMN IF EXISTS deleted_at;
