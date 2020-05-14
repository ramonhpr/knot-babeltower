package client

import (
	"log"

	"github.com/CESARBR/knot-babeltower/internal/config"
	"github.com/streadway/amqp"
)

// SimpleClient is a simple AMQP client
type SimpleClient interface {
	Connect(config.RabbitMQ) error
	Send(string, string, []byte, map[string]interface{}) error
	Subscribe(string, string, map[string]interface{}) (chan []byte, error)
	GetQueueName() string
}

type simpleClient struct {
	channel   *amqp.Channel
	queueName string
}

// NewSimpleSender constructs a new sender
func NewSimpleSender() SimpleClient {
	return &simpleClient{}
}

func (s *simpleClient) Connect(config config.RabbitMQ) error {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		log.Printf("Unable to connect: %s", err)
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Unable to create channel: %s", err)
		return err
	}
	s.channel = ch
	return nil
}

func (s *simpleClient) Send(exchange, key string, body []byte, headers map[string]interface{}) error {

	return s.channel.Publish(
		exchange, // exchange
		key,      // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			Headers:     headers,
			ContentType: "text/plain",
			Body:        body,
		})
}

func (s *simpleClient) GetQueueName() string {
	return s.queueName
}

func (s *simpleClient) Subscribe(exchange, key string, headers map[string]interface{}) (chan []byte, error) {
	queue, err := s.channel.QueueDeclare("", false, false, true, true, headers)
	if err != nil {
		return nil, err
	}

	s.queueName = queue.Name
	err = s.channel.QueueBind(queue.Name, key, exchange, true, nil)
	if err != nil {
		return nil, err
	}

	chanDelivery, err := s.channel.Consume(queue.Name, "", true, true, true, true, nil)
	if err != nil {
		return nil, err
	}
	chanRet := make(chan []byte)
	go func(deliveries <-chan amqp.Delivery, outChan chan []byte) {
		for d := range deliveries {
			outChan <- d.Body
		}
	}(chanDelivery, chanRet)
	return chanRet, nil
}
