package settings

// UpdateInput contains the payload to update user settings.
type UpdateInput struct {
	PreferredCurrency string `json:"preferred_currency"`
	Timezone          string `json:"timezone"`
	DateFormat        string `json:"date_format"`
	WeekStartDay      string `json:"week_start_day"`
}
