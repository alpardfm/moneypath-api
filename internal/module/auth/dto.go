package auth

// RegisterInput contains the user registration payload.
type RegisterInput struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

// LoginInput contains the login payload.
type LoginInput struct {
	EmailOrUsername string `json:"email_or_username"`
	Password        string `json:"password"`
}

// AuthResult contains the authenticated user and token.
type AuthResult struct {
	Token string
	User  *User
}
