package network

import (
	"github.com/cenkalti/backoff/v4"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
	"github.com/streadway/amqp"
)

// IQueueService is the interface that handles remote queue service
type IQueueService interface {
	Start(started chan bool)
	Stop()
	GetPublisher() IPublisher
	GetSubscriber() ISubscriber
}

// IPublisher is the interface to publish message
type IPublisher interface {
	PublishPersistentMessage(exchange, key string, body []byte) error
}

// ISubscriber is the interface to receive messages
type ISubscriber interface {
	OnMessage(msgChan chan InMsg, queueName, exchangeName, key string) error
}

// Amqp handles the connection, queues and exchanges declared
type Amqp struct {
	url     string
	logger  logging.Logger
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   *amqp.Queue
}

// InMsg represents the message received from the AMQP broker
type InMsg struct {
	Exchange   string
	RoutingKey string
	Headers    map[string]interface{}
	Body       []byte
}

// NewAmqp constructs the AMQP connection handler
func NewAmqp(url string, logger logging.Logger) *Amqp {
	return &Amqp{url, logger, nil, nil, nil}
}

// GetPublisher returns the publisher
func (a *Amqp) GetPublisher() IPublisher {
	return a
}

// GetSubscriber returns the subscriber
func (a *Amqp) GetSubscriber() ISubscriber {
	return a
}

// Start starts the handler
func (a *Amqp) Start(started chan bool) {
	err := backoff.Retry(a.connect, backoff.NewExponentialBackOff())
	if err != nil {
		a.logger.Error(err)
		started <- false
		return
	}

	go a.notifyWhenClosed(started)
	started <- true
}

// Stop closes the connection started
func (a *Amqp) Stop() {
	if a.conn != nil && !a.conn.IsClosed() {
		a.conn.Close()
	}

	if a.channel != nil {
		a.channel.Close()
	}

	a.logger.Debug("AMQP handler stopped")
}

// PublishPersistentMessage sends a persistent message to RabbitMQ
func (a *Amqp) PublishPersistentMessage(exchange, key string, body []byte) error {
	err := a.declareExchange(exchange)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.channel.Publish(
		exchange,
		key,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    amqp.Persistent,
			Priority:        0,
		},
	)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	return nil
}

// OnMessage receive messages and put them on channel
func (a *Amqp) OnMessage(msgChan chan InMsg, queueName, exchangeName, key string) error {
	err := a.declareExchange(exchangeName)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.declareQueue(queueName)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	err = a.channel.QueueBind(
		queueName,
		key,
		exchangeName,
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	deliveries, err := a.channel.Consume(
		queueName,
		"",    // consumerTag
		true,  // noAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	go convertDeliveryToInMsg(deliveries, msgChan)

	return nil
}

func (a *Amqp) connect() error {
	conn, err := amqp.Dial(a.url)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	a.conn = conn

	channel, err := a.conn.Channel()
	if err != nil {
		a.logger.Error(err)
		return err
	}

	a.logger.Debug("AMQP handler connected")
	a.channel = channel

	return nil
}

func (a *Amqp) notifyWhenClosed(started chan bool) {
	errReason := <-a.conn.NotifyClose(make(chan *amqp.Error))
	a.logger.Infof("AMQP connection closed: %s", errReason)
	started <- false
	if errReason != nil {
		err := backoff.Retry(a.connect, backoff.NewExponentialBackOff())
		if err != nil {
			a.logger.Error(err)
			started <- false
			return
		}

		go a.notifyWhenClosed(started)
		started <- true
	}
}

func (a *Amqp) declareExchange(name string) error {
	return a.channel.ExchangeDeclare(
		name,
		amqp.ExchangeTopic, // type
		true,               // durable
		false,              // delete when complete
		false,              // internal
		false,              // noWait
		nil,                // arguments
	)
}

func (a *Amqp) declareQueue(name string) error {
	queue, err := a.channel.QueueDeclare(
		name,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)

	a.queue = &queue
	return err
}

func convertDeliveryToInMsg(deliveries <-chan amqp.Delivery, outMsg chan InMsg) {
	for d := range deliveries {
		outMsg <- InMsg{d.Exchange, d.RoutingKey, d.Headers, d.Body}
	}
}
