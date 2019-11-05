package entities

// User User represents the user domain model
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
