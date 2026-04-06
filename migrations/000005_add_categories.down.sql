DROP INDEX IF EXISTS idx_mutations_category_id_happened_at;

ALTER TABLE mutations
    DROP COLUMN IF EXISTS category_id;

DROP INDEX IF EXISTS idx_categories_user_id_deleted_at;
DROP INDEX IF EXISTS idx_categories_user_id_type_active;
DROP INDEX IF EXISTS idx_categories_user_id_name_type_active_unique;

DROP TABLE IF EXISTS categories;
