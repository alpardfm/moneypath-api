package category

import "time"

// Category is a user-owned classification for mutation records.
type Category struct {
	ID        string
	UserID    string
	Name      string
	Type      string
	IsActive  bool
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
