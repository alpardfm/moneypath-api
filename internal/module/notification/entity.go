package notification

import "time"

// Item represents one in-app notification entry.
type Item struct {
	Type         string     `json:"type"`
	Severity     string     `json:"severity"`
	Title        string     `json:"title"`
	Message      string     `json:"message"`
	ResourceID   string     `json:"resource_id,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"`
	OccurredAt   *time.Time `json:"occurred_at,omitempty"`
}

// Report contains the current notification feed for the authenticated user.
type Report struct {
	Items []Item `json:"items"`
}
