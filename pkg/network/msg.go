package network

// InMsg received from RabbitMQ
type InMsg struct {
	Exchange   string
	RoutingKey string
	Body       []byte
}
