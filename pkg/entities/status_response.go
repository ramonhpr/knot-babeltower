package entities

// StatusResponse represents the response to be sent to the request
type StatusResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
