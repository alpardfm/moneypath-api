package category

// CreateInput contains the create category payload.
type CreateInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// UpdateInput contains the update category payload.
type UpdateInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
