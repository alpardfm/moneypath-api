CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category_type TEXT NOT NULL CHECK (category_type IN ('masuk', 'keluar')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_categories_user_id_name_type_active_unique
    ON categories (user_id, LOWER(name), category_type)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_categories_user_id_type_active
    ON categories (user_id, category_type, is_active);

CREATE INDEX idx_categories_user_id_deleted_at
    ON categories (user_id, deleted_at);

ALTER TABLE mutations
    ADD COLUMN category_id UUID REFERENCES categories(id) ON DELETE RESTRICT;

CREATE INDEX idx_mutations_category_id_happened_at
    ON mutations (category_id, happened_at DESC);
