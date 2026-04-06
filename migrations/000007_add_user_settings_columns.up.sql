ALTER TABLE users
    ADD COLUMN preferred_currency TEXT NOT NULL DEFAULT 'IDR',
    ADD COLUMN timezone TEXT NOT NULL DEFAULT 'Asia/Jakarta',
    ADD COLUMN date_format TEXT NOT NULL DEFAULT 'YYYY-MM-DD' CHECK (date_format IN ('YYYY-MM-DD', 'DD-MM-YYYY', 'MM-DD-YYYY')),
    ADD COLUMN week_start_day TEXT NOT NULL DEFAULT 'monday' CHECK (week_start_day IN ('monday', 'sunday'));
