ALTER TABLE mutations
    ADD COLUMN debt_action TEXT NOT NULL DEFAULT 'none'
    CHECK (debt_action IN ('none', 'payment', 'borrow_existing', 'borrow_new'));
