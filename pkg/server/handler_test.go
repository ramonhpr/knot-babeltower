package server

import (
	"testing"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
	"github.com/CESARBR/knot-babeltower/pkg/network"
	"github.com/CESARBR/knot-babeltower/pkg/thing/controllers"
	"github.com/CESARBR/knot-babeltower/pkg/thing/mocks"
)

type FakeSubscriber struct {
}

func (f *FakeSubscriber) OnMessage(msgChan chan network.InMsg, queueName string, exchangeName string, key string) error {
	return nil
}

type FakePublisher struct {
}

func (f *FakePublisher) PublishPersistentMessage(exchange string, key string, body []byte) error {
	panic("not implemented")
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
}

func (f *FakeController) Register(body []byte, authorizationHeader string) error {
	panic("not implemented")
}

func (f *FakeController) Unregister(body []byte, authorizationHeader string) error {
	panic("not implemented")
}

func (f *FakeController) UpdateSchema(body []byte, authorizationHeader string) error {
	panic("not implemented")
}

func (f *FakeController) AuthDevice(body []byte, authorization string) error {
	panic("not implemented")
}

func (f *FakeController) ListDevices(authorization string) error {
	panic("not implemented")
}

func (f *FakeController) PublishData(body []byte, authorization string) error {
	panic("not implemented")
}

func (f *FakeController) RequestData(body []byte, authorization string) error {
	panic("not implemented")
}

func (f *FakeController) UpdateData(body []byte, authorization string) error {
	panic("not implemented")
}

func TestMsgHandler_Start(t *testing.T) {
	type fields struct {
		logger          logging.Logger
		amqp            network.IQueueService
		thingController controllers.IController
	}
	type args struct {
		started chan bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"when channel nil start should return error",
			fields{
				&mocks.FakeLogger{},
				&FakeQueue{},
				&FakeController{},
			},
			args{
				nil,
			},
			true,
		},
		// TODO: Add test cases.
		// test channel started nil DONE
		// received correct message
		// called correct controller function
		// called correct connector function
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MsgHandler{
				logger:          tt.fields.logger,
				amqp:            tt.fields.amqp,
				thingController: tt.fields.thingController,
			}
			err := mc.Start(tt.args.started)
			if (err != nil) != tt.wantErr {
				t.Errorf("msgClientPublisher.SendRegisteredDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
