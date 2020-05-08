package server

import (
	"testing"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
	"github.com/CESARBR/knot-babeltower/pkg/mocks"
	"github.com/CESARBR/knot-babeltower/pkg/network"
	"github.com/CESARBR/knot-babeltower/pkg/thing/controllers"
)

func TestStart(t *testing.T) {
	type fields struct {
		logger          logging.Logger
		amqp            network.AmqpReceiver
		thingController controllers.ThingController
	}
	type args struct {
		started chan bool
		msgChan chan network.InMsg
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		expectedErr bool
	}{
		{
			"when started channel not provided an error should be returned",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				nil,
				make(chan network.InMsg),
			},
			true,
		},
		{
			"when msg channel not provided an error should be returned",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
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
			if (err != nil) != tt.expectedErr {
				t.Errorf("msgClientPublisher.SendRegisteredDevice() error = %v, wantErr %v", err, tt.expectedErr)
			}
		})
	}
}

func TestOnMsgReceived(t *testing.T) {
	type fields struct {
		logger          *mocks.FakeLogger
		amqp            network.AmqpReceiver
		thingController *mocks.FakeController
	}
	type args struct {
		started chan bool
		msgChan chan network.InMsg
		msg     network.InMsg
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		mockArgs map[string]string
	}{
		{
			"happy path device exchange without RPC should return no error",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{
					Exchange:   exchangeDevices,
					RoutingKey: bindingKeyRegisterDevice,
					Body:       []byte{1, 2, 3},
					Headers: map[string]interface{}{
						"Authorization": "test-token",
					},
				},
			},
			false,
			map[string]string{
				bindingKeyRegisterDevice:   "Register",
				bindingKeyUnregisterDevice: "Unregister",
				bindingKeyRequestData:      "RequestData",
				bindingKeyUpdateData:       "UpdateData",
				bindingKeySchemaSent:       "UpdateSchema",
			},
		},
		{
			"happy path device exchange with RPC should return no error",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{
					Exchange:   exchangeDevices,
					RoutingKey: bindingKeyRequestData,
					Body:       []byte{1, 2, 3},
					Headers: map[string]interface{}{
						"Authorization":  "test-token",
						"correlation_id": "test-corrId",
						"reply_to":       "test-reply_to",
					},
				},
			},
			false,
			map[string]string{
				bindingKeyAuthDevice:  "AuthDevice",
				bindingKeyListDevices: "ListDevices",
			},
		},
		{
			"happy path when send data to fanout exchange should ignore routing and call correct function",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{
					Exchange:   exchangeDataSent,
					RoutingKey: "any.key",
					Body:       []byte{1, 2, 3},
					Headers: map[string]interface{}{
						"Authorization": "test-token",
					},
				},
			},
			false,
			map[string]string{
				"any.key": "PublishData",
			},
		},
		{
			"missing correlation id on device exchange with RPC should return missing correlation ID",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{
					Exchange:   exchangeDevices,
					RoutingKey: bindingKeyListDevices,
					Body:       []byte{1, 2, 3},
					Headers: map[string]interface{}{
						"Authorization": "test-token",
						"reply_to":      "test-reply_to",
					},
				},
			},
			true,
			nil,
		},
		{
			"missing reply_to on device exchange with RPC should return missing reply to",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{
					Exchange:   exchangeDevices,
					RoutingKey: bindingKeyListDevices,
					Body:       []byte{1, 2, 3},
					Headers: map[string]interface{}{
						"Authorization":  "test-token",
						"correlation_id": "test-corrId",
					},
				},
			},
			true,
			nil,
		},
		{
			"empty header should return missing authorization token",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: exchangeDevices, RoutingKey: bindingKeyRegisterDevice, Body: []byte{1, 2, 3}, Headers: map[string]interface{}{}},
			},
			true,
			nil,
		},
		{
			"unexpected exchange should return operation unsuported",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: "test", RoutingKey: bindingKeyRegisterDevice, Body: []byte{1, 2, 3}, Headers: map[string]interface{}{"Authorization": "test-token"}},
			},
			true,
			nil,
		},
		{
			"when message is from client and unexpected routing key should return operation unsuported",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{
					Exchange:   exchangeDevices,
					RoutingKey: "key",
					Body:       []byte{1, 2, 3},
					Headers: map[string]interface{}{
						"Authorization": "test-token",
					},
				},
			},
			true,
			nil,
		},
		{
			"when header is not provided should return an error",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{Exchange: exchangeDevices, RoutingKey: bindingKeyRegisterDevice, Body: []byte{1, 2, 3}, Headers: nil},
			},
			true,
			nil,
		},
		{
			"when body is not provided should return an error",
			fields{
				&mocks.FakeLogger{},
				&mocks.FakeAmqpReceiver{},
				&mocks.FakeController{},
			},
			args{
				make(chan bool, 1),
				make(chan network.InMsg, 10),
				network.InMsg{
					Exchange:   exchangeDevices,
					RoutingKey: bindingKeyRegisterDevice,
					Body:       nil,
					Headers: map[string]interface{}{
						"Authorization": "test-token",
					},
				},
			},
			true,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MsgHandler{
				logger:          tt.fields.logger,
				amqp:            tt.fields.amqp,
				thingController: tt.fields.thingController,
			}
			if tt.mockArgs != nil {
				for key, value := range tt.mockArgs {
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
