package wallet

// CreateInput contains the create wallet payload.
type CreateInput struct {
	Name string `json:"name"`
}

// UpdateInput contains the update wallet payload.
type UpdateInput struct {
	Name string `json:"name"`
}
