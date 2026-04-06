ALTER TABLE users
    DROP COLUMN IF EXISTS week_start_day,
    DROP COLUMN IF EXISTS date_format,
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS preferred_currency;
