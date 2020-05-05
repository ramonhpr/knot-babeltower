package server

import (
	"errors"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
	"github.com/CESARBR/knot-babeltower/pkg/network"
	"github.com/CESARBR/knot-babeltower/pkg/thing/controllers"
)

const (
	queueNameFogIn           = "fogIn-messages"
	exchangeFogIn            = "fogIn"
	queueNameConnOut         = "connOut-messages"
	exchangeConnOut          = "connOut"
	bindingKeyDevice         = "device.*"
	bindingKeyPublishData    = "data.publish"
	bindingKeyDataCommands   = "data.*"
	bindingKeyDeviceCommands = "device.cmd.*"
	bindingKeySchema         = "schema.*"
)

// MsgHandler handle messages received from a service
type MsgHandler struct {
	logger          logging.Logger
	amqp            network.IQueueService
	thingController controllers.IController
}

// NewMsgHandler creates a new MsgHandler instance with the necessary dependencies
func NewMsgHandler(logger logging.Logger, amqp network.IQueueService, thingController *controllers.ThingController) *MsgHandler {
	return &MsgHandler{logger, amqp, thingController}
}

// Start starts to listen messages
func (mc *MsgHandler) Start(started chan bool) error {
	return mc.start(started, make(chan network.InMsg))
}

func (mc *MsgHandler) start(started chan bool, msgChan chan network.InMsg) error {
	mc.logger.Debug("message handler started")
	if started == nil {
		return errors.New("missing channel")
	}
	if msgChan == nil {
		return errors.New("missing msgChan")
	}
	err := mc.subscribeToMessages(msgChan)
	if err != nil {
		mc.logger.Error(err)
		started <- false
		return err
	}

	go func() {
		for {
			err := mc.onMsgReceived(msgChan)
			if err != nil {
				mc.logger.Error(err)
			}
		}
	}()

	started <- true
	return nil
}

// Stop stops to listen for messages
func (mc *MsgHandler) Stop() {
	mc.logger.Debug("message handler stopped")
}

func (mc *MsgHandler) subscribeToMessages(msgChan chan network.InMsg) error {
	var err error
	subscribe := func(msgChan chan network.InMsg, queueName, exchange, key string) {
		if err != nil {
			return
		}
		err = mc.amqp.GetSubscriber().OnMessage(msgChan, queueName, exchange, key)
	}

	// Subscribe to messages received from any client
	subscribe(msgChan, queueNameFogIn, exchangeFogIn, bindingKeyDevice)
	subscribe(msgChan, queueNameFogIn, exchangeFogIn, bindingKeySchema)
	subscribe(msgChan, queueNameFogIn, exchangeFogIn, bindingKeyDeviceCommands)
	subscribe(msgChan, queueNameFogIn, exchangeFogIn, bindingKeyPublishData)

	// Subscribe to messages received from the connector service
	subscribe(msgChan, queueNameConnOut, exchangeConnOut, bindingKeyDataCommands)
	subscribe(msgChan, queueNameConnOut, exchangeConnOut, bindingKeyDevice)

	return err
}

func (mc *MsgHandler) onMsgReceived(msgChan chan network.InMsg) error {
	msg := <-msgChan
	var err error
	mc.logger.Infof("exchange: %s, routing key: %s", msg.Exchange, msg.RoutingKey)
	mc.logger.Infof("message received: %s", string(msg.Body))

	token, ok := msg.Headers["Authorization"].(string)
	if !ok {
		return errors.New("authorization token not provided")
	}

	if msg.Exchange == exchangeFogIn {
		err = mc.handleClientMessages(msg, token)
	} else if msg.Exchange == exchangeConnOut {
		err = mc.handleConnectorMessages(msg, token)
	} else {
		err = errors.New("unexpected exchange received")
	}

	if err != nil {
		return err
	}

	return nil
}

func (mc *MsgHandler) handleClientMessages(msg network.InMsg, token string) error {

	switch msg.RoutingKey {
	case "device.register":
		return mc.thingController.Register(msg.Body, token)
	case "device.unregister":
		return mc.thingController.Unregister(msg.Body, token)
	case "schema.update":
		return mc.thingController.UpdateSchema(msg.Body, token)
	case "device.cmd.auth":
		return mc.thingController.AuthDevice(msg.Body, token)
	case "device.cmd.list":
		return mc.thingController.ListDevices(token)
	case "data.publish":
		return mc.thingController.PublishData(msg.Body, token)
	default:
		return errors.New("unexpected routing key")
	}
}

func (mc *MsgHandler) handleConnectorMessages(msg network.InMsg, token string) error {

	switch msg.RoutingKey {
	case "data.request":
		return mc.thingController.RequestData(msg.Body, token)
	case "data.update":
		return mc.thingController.UpdateData(msg.Body, token)
	case "device.registered":
		// Ignore message
		break
	default:
		return errors.New("unexpected routing key")
	}

	return nil
}
