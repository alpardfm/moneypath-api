package debt

// CreateInput contains the create debt payload.
type CreateInput struct {
	Name          string  `json:"name"`
	Principal     string  `json:"principal_amount"`
	TenorValue    *int    `json:"tenor_value"`
	TenorUnit     *string `json:"tenor_unit"`
	PaymentAmount *string `json:"payment_amount"`
	Note          *string `json:"note"`
}

// UpdateInput contains the update debt metadata payload.
type UpdateInput struct {
	Name          string  `json:"name"`
	TenorValue    *int    `json:"tenor_value"`
	TenorUnit     *string `json:"tenor_unit"`
	PaymentAmount *string `json:"payment_amount"`
	Note          *string `json:"note"`
}
