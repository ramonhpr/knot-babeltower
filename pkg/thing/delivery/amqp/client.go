package amqp

import (
	"encoding/json"
	"fmt"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
	"github.com/CESARBR/knot-babeltower/pkg/network"
	"github.com/CESARBR/knot-babeltower/pkg/thing/entities"
)

const (
	exchangeDevices           = "device"
	exchangeDevicesType       = "direct"
	exchangeDataPublished     = "data.published"
	exchangeDataPublishedType = "fanout"
	registerOutKey            = "device.registered"
	unregisterOutKey          = "device.unregistered"
	schemaOutKey              = "device.schema.updated"
	updateDataKey             = "data.update"
	requestDataKey            = "data.request"
)

// Publisher provides methods to send events to the clients
type Publisher interface {
	PublishRegisteredDevice(thingID, name, token string, err error) error
	PublishUnregisteredDevice(thingID string, err error) error
	PublishUpdatedSchema(thingID string, schema []entities.Schema, err error) error
	PublishUpdateData(thingID string, data []entities.Data) error
	PublishRequestData(thingID string, sensorIds []int) error
	PublishPublishedData(thingID, token string, data []entities.Data) error
}

// Sender represents the operations to send commands response
type Sender interface {
	SendAuthResponse(thingID, replyTo, corrID string, err error) error
	SendListResponse(things []*entities.Thing, replyTo, corrID string, err error) error
}

// msgClientPublisher handle messages received from a service
type msgClientPublisher struct {
	logger logging.Logger
	amqp   network.AmqpSender
}

// commandSender handle messages received from a service
type commandSender struct {
	logger logging.Logger
	amqp   network.AmqpSender
}

// NewMsgClientPublisher constructs the msgClientPublisher
func NewMsgClientPublisher(logger logging.Logger, amqp network.AmqpSender) Publisher {
	return &msgClientPublisher{logger, amqp}
}

// NewCommandSender creates a new commandSender instance
func NewCommandSender(logger logging.Logger, amqp network.AmqpSender) Sender {
	return &commandSender{logger, amqp}
}

// PublishRegisteredDevice publishes the registered device's credentials to the device registration queue
func (mp *msgClientPublisher) PublishRegisteredDevice(thingID, name, token string, err error) error {
	mp.logger.Debug("sending registered message")
	errMsg := getErrMsg(err)
	resp := &network.DeviceRegisteredResponse{ID: thingID, Name: name, Token: token, Error: errMsg}
	msg, err := json.Marshal(resp)
	if err != nil {
		mp.logger.Error(err)
		return err
	}

	return mp.amqp.PublishPersistentMessage(exchangeDevices, exchangeDevicesType, registerOutKey, msg, nil)
}

// PublishUnregisteredDevice publishes the unregistered device's id and error message to the device unregistered queue
func (mp *msgClientPublisher) PublishUnregisteredDevice(thingID string, err error) error {
	mp.logger.Debug("sending unregistered message")
	errMsg := getErrMsg(err)
	resp := &network.DeviceUnregisteredResponse{ID: thingID, Error: errMsg}
	msg, err := json.Marshal(resp)
	if err != nil {
		mp.logger.Error(err)
		return err
	}

	return mp.amqp.PublishPersistentMessage(exchangeDevices, exchangeDevicesType, unregisterOutKey, msg, nil)
}

// PublishUpdatedSchema sends the updated schema response
func (mp *msgClientPublisher) PublishUpdatedSchema(thingID string, schema []entities.Schema, err error) error {
	errMsg := getErrMsg(err)
	resp := &network.SchemaUpdatedResponse{ID: thingID, Schema: schema, Error: errMsg}
	msg, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return mp.amqp.PublishPersistentMessage(exchangeDevices, exchangeDevicesType, schemaOutKey, msg, nil)
}

// PublishRequestData sends request data command
func (mp *msgClientPublisher) PublishRequestData(thingID string, sensorIds []int) error {
	resp := &network.DataRequest{ID: thingID, SensorIds: sensorIds}
	msg, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	routingKey := "device." + thingID + "." + requestDataKey
	return mp.amqp.PublishPersistentMessage(exchangeDevices, exchangeDevicesType, routingKey, msg, nil)
}

// PublishUpdateData send update data command
func (mp *msgClientPublisher) PublishUpdateData(thingID string, data []entities.Data) error {
	resp := &network.DataUpdate{ID: thingID, Data: data}
	msg, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("message parsing error: %w", err)
	}

	routingKey := "device." + thingID + "." + updateDataKey
	return mp.amqp.PublishPersistentMessage(exchangeDevices, exchangeDevicesType, routingKey, msg, nil)
}

// SendAuthResponse sends the auth thing status response
func (cs *commandSender) SendAuthResponse(thingID string, replyTo, corrID string, err error) error {
	errMsg := getErrMsg(err)
	resp := &network.DeviceAuthResponse{ID: thingID, Error: errMsg}
	headers := map[string]interface{}{
		"correlation_id": corrID,
	}
	msg, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return cs.amqp.PublishPersistentMessage(exchangeDevices, exchangeDevicesType, replyTo, msg, headers)
}

// SendListResponse sends the list devices command response
func (cs *commandSender) SendListResponse(things []*entities.Thing, replyTo, corrID string, err error) error {
	errMsg := getErrMsg(err)
	resp := &network.DeviceListResponse{Things: things, Error: errMsg}
	headers := map[string]interface{}{
		"correlation_id": corrID,
	}
	msg, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return cs.amqp.PublishPersistentMessage(exchangeDevices, exchangeDevicesType, replyTo, msg, headers)
}

// PublishPublishedData send update data command
func (mp *msgClientPublisher) PublishPublishedData(thingID, token string, data []entities.Data) error {
	resp := &network.DataSent{ID: thingID, Data: data}
	headers := map[string]interface{}{
		"Authorization": token,
	}
	msg, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("message parsing error: %w", err)
	}

	return mp.amqp.PublishPersistentMessage(exchangeDataPublished, exchangeDataPublishedType, "", msg, headers)
}

func getErrMsg(err error) *string {
	if err != nil {
		msg := err.Error()
		return &msg
	}
	return nil
}
