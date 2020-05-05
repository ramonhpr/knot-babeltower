package server

import (
	"errors"
	"testing"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
	"github.com/CESARBR/knot-babeltower/pkg/network"
	"github.com/CESARBR/knot-babeltower/pkg/thing/controllers"
	"github.com/CESARBR/knot-babeltower/pkg/thing/mocks"
	"github.com/stretchr/testify/mock"
)

type FakeSubscriber struct {
}

func (f *FakeSubscriber) OnMessage(msgChan chan network.InMsg, queueName string, exchangeName string, key string) error {
	return nil
}

type FakePublisher struct {
}

func (f *FakePublisher) PublishPersistentMessage(exchange string, key string, body []byte) error {
	return nil
}

type FakeQueue struct {
	pub *FakePublisher
	sub *FakeSubscriber
}

func (f *FakeQueue) Start(started chan bool) {
	f.pub = &FakePublisher{}
	f.sub = &FakeSubscriber{}
}

func (f *FakeQueue) Stop() {
	panic("not implemented")
}

func (f *FakeQueue) GetPublisher() network.IPublisher {
	return f.pub
}

func (f *FakeQueue) GetSubscriber() network.ISubscriber {
	return f.sub
}

type FakeController struct {
	mock.Mock
}

func (f *FakeController) Register(body []byte, authorizationHeader string) error {
	if len(body) == 0 {
		return errors.New("empty body")
	}
	f.Called()
	return nil
}

func (f *FakeController) Unregister(body []byte, authorizationHeader string) error {
	f.Called()
	return nil
}

func (f *FakeController) UpdateSchema(body []byte, authorizationHeader string) error {
	f.Called()
	return nil
}

func (f *FakeController) AuthDevice(body []byte, authorization string) error {
	f.Called()
	return nil
}

func (f *FakeController) ListDevices(authorization string) error {
	f.Called()
	return nil
}

func (f *FakeController) PublishData(body []byte, authorization string) error {
	f.Called()
	return nil
}

func (f *FakeController) RequestData(body []byte, authorization string) error {
	f.Called()
	return nil
}

func (f *FakeController) UpdateData(body []byte, authorization string) error {
	f.Called()
	return nil
}

func TestMsgHandler_Start(t *testing.T) {
	type fields struct {
		logger          logging.Logger
		amqp            network.IQueueService
		thingController controllers.IController
	}
	type args struct {
		started chan bool
		msgChan chan network.InMsg
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"when started channel nil start should return error",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				nil,
				make(chan network.InMsg),
			},
			true,
		},
		{
			"when msg channel nil start should return error",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				make(chan bool),
				nil,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MsgHandler{
				logger:          tt.fields.logger,
				amqp:            tt.fields.amqp,
				thingController: tt.fields.thingController,
			}
			err := mc.start(tt.args.started, tt.args.msgChan)
			if (err != nil) != tt.wantErr {
				t.Errorf("msgClientPublisher.SendRegisteredDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgHandler_onMsgReceived(t *testing.T) {
	type fields struct {
		logger          *mocks.FakeLogger
		amqp            network.IQueueService
		thingController *FakeController
	}
	type args struct {
		started chan bool
		msgChan chan network.InMsg
		msg     network.InMsg
	}
	tests := []struct {
		name                     string
		fields                   fields
		args                     args
		wantErr                  bool
		mapKeyToControllerMethod map[string]string
	}{
		{
			"happy path fog exchange should return no error",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: exchangeFogIn, RoutingKey: "device.register", Body: []byte{1, 2, 3}, Headers: map[string]interface{}{"Authorization": "test-token"}},
			},
			false,
			map[string]string{
				"device.register":   "Register",
				"device.unregister": "Unregister",
				"schema.update":     "UpdateSchema",
				"device.cmd.auth":   "AuthDevice",
				"device.cmd.list":   "ListDevices",
				"data.publish":      "PublishData",
			},
		},
		{
			"happy path connector exchange should return no error",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: exchangeConnOut, RoutingKey: "data.request", Body: []byte{1, 2, 3}, Headers: map[string]interface{}{"Authorization": "test-token"}},
			},
			false,
			map[string]string{"data.request": "RequestData",
				"data.update":       "UpdateData",
				"device.registered": "",
			},
		},
		{
			"empty header should return missing authorization token",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: exchangeFogIn, RoutingKey: "device.register", Body: []byte{1, 2, 3}, Headers: map[string]interface{}{}},
			},
			true,
			nil,
		},
		{
			"unexpected exchange should return operation unsuported",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: "test", RoutingKey: "device.register", Body: []byte{1, 2, 3}, Headers: map[string]interface{}{"Authorization": "test-token"}},
			},
			true,
			nil,
		},
		{
			"when message is from client and unexpected routing key should return operation unsuported",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: exchangeFogIn, RoutingKey: "key", Body: []byte{1, 2, 3}, Headers: map[string]interface{}{"Authorization": "test-token"}},
			},
			true,
			nil,
		},
		{
			"when message is from connector and unexpected routing key should return operation unsuported",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: exchangeConnOut, RoutingKey: "key", Body: []byte{1, 2, 3}, Headers: map[string]interface{}{"Authorization": "test-token"}},
			},
			true,
			nil,
		},
		{
			"when header is nil should return exception",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: exchangeFogIn, RoutingKey: "device.register", Body: []byte{1, 2, 3}, Headers: nil},
			},
			true,
			nil,
		},
		{
			"when body is nil should return exception",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: exchangeFogIn, RoutingKey: "device.register", Body: nil, Headers: map[string]interface{}{"Authorization": "test-token"}},
			},
			true,
			nil,
		},
		// called correct controller function
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MsgHandler{
				logger:          tt.fields.logger,
				amqp:            tt.fields.amqp,
				thingController: tt.fields.thingController,
			}
			if tt.mapKeyToControllerMethod != nil {
				for key, value := range tt.mapKeyToControllerMethod {
					if len(value) > 0 {
						tt.fields.thingController.On(value).Once()
					}
					tt.args.msg.RoutingKey = key
					tt.args.msgChan <- tt.args.msg
					err := mc.onMsgReceived(tt.args.msgChan)
					if (err != nil) != tt.wantErr {
						t.Errorf("msgClientPublisher.SendRegisteredDevice() error = %v, wantErr %v", err, tt.wantErr)
					}
					if len(value) > 0 {
						tt.fields.thingController.AssertExpectations(t)
					}
				}
			} else {
				tt.args.msgChan <- tt.args.msg
				err := mc.onMsgReceived(tt.args.msgChan)
				if (err != nil) != tt.wantErr {
					t.Errorf("msgClientPublisher.SendRegisteredDevice() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
