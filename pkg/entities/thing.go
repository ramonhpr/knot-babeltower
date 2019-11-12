package entities

type Thing struct {
	Name   string                 `json:"name"`
	Id     string                 `json:"id"`
	Schema map[string]interface{} `json:"schema"`
}
