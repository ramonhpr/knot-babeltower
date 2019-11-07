package entities

// ConflictResponse represents the HTTP code 409 Conflict
type ConflictResponse struct{}

func (err ConflictResponse) Error() string {
	return "Error 409"
}
