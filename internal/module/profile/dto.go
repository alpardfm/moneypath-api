package profile

// UpdateProfileInput contains the payload to update the current user profile.
type UpdateProfileInput struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

// ChangePasswordInput contains the payload to change the current password.
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
