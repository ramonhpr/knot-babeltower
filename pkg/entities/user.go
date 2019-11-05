package entities

// User represents the user data from update request
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
