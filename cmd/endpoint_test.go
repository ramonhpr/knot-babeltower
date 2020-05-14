package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	cli "github.com/CESARBR/knot-babeltower/cmd/client"
	"github.com/CESARBR/knot-babeltower/internal/config"
	"github.com/CESARBR/knot-babeltower/pkg/mocks"
	"github.com/CESARBR/knot-babeltower/pkg/network"
	thingEntities "github.com/CESARBR/knot-babeltower/pkg/thing/entities"
	"github.com/CESARBR/knot-babeltower/pkg/user/delivery/http"
	"github.com/CESARBR/knot-babeltower/pkg/user/entities"
	"github.com/stretchr/testify/assert"
)

var (
	quitMain chan bool
	sender   cli.SimpleClient
	rpc      cli.RPCService
	token    string
)

func getToken(config config.Users) (string, error) {
	proxy := http.NewUserProxy(&mocks.FakeLogger{}, config.Hostname, config.Port)
	user := entities.User{Email: "test@test.com", Password: "12345678"}
	_ = proxy.Create(user)

	return proxy.CreateToken(user)
}

func SetupSuite() error {
	quitMain = make(chan bool, 1)
	started := make(chan bool, 1)
	config := config.GetDefaultConfig()
	go Main(config, quitMain, started)
	select {
	case <-started:
		break
	case <-time.After(20 * time.Second):
		return errors.New("timeout waiting broker startup")
	}

	var err error
	token, err = getToken(config.Users)
	if err != nil {
		return err
	}
	sender = cli.NewSimpleSender()
	rpc = cli.NewSimpleService(config.RabbitMQ, token)
	return sender.Connect(config.RabbitMQ)
}

func TearDownSuite() error {
	quitMain <- true
	return nil
}

// Happy path client send and receive response
// Happy path publisher send and subscriber receive
// Two clients send and each one receives the response
// One publisher and two subscribers
// Two publishers and one subscriber should receive both messages
// Test invalid configuration should return error/panic

func Test_HappyPath_PubSub(t *testing.T) {
	tests := []struct {
		name         string
		exchange     string
		key          string
		respExchange string
		respKey      string
		msg          interface{}
		msgResp      interface{}
	}{
		{
			"Happy path client send register then receive response",
			"device",
			"device.register",
			"device",
			"device.registered",
			network.DeviceRegisterRequest{ID: "123", Name: "testThing"},
			network.DeviceRegisteredResponse{},
		},
		{
			"Happy path client send schema then receive response",
			"device",
			"device.schema.sent",
			"device",
			"device.schema.updated",
			network.SchemaUpdateRequest{
				ID: "123",
				Schema: []thingEntities.Schema{
					{SensorID: 1, ValueType: 2, Unit: 1, TypeID: 13, Name: "testSensor"},
				},
			},
			network.SchemaUpdatedResponse{},
		},
		{
			"Happy path client send data then receive event published",
			"data.sent",
			"",
			"data.published",
			"",
			network.DataSent{
				ID: "123",
				Data: []thingEntities.Data{
					{SensorID: 1, Value: 12.5},
				},
			},
			network.DataSent{},
		},
		{
			"Happy path client send request data then receive response",
			"device",
			"data.request",
			"device",
			"device.123.data.request",
			network.DataRequest{
				ID:        "123",
				SensorIds: []int{1},
			},
			network.DataRequest{},
		},
		{
			"Happy path client send update data then receive response",
			"device",
			"data.update",
			"device",
			"device.123.data.update",
			network.DataUpdate{
				ID: "123",
				Data: []thingEntities.Data{
					{SensorID: 1, Value: 20.7},
				}},
			network.DataUpdate{},
		},
		{
			"Happy path client send unregister then receive response",
			"device",
			"device.unregister",
			"device",
			"device.unregistered",
			network.DeviceUnregisterRequest{ID: "123"},
			network.DeviceUnregisteredResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.msg)
			if err != nil {
				assert.FailNow(t, err.Error())
			}

			t.Log("send to: ", tt.exchange, tt.key, sender)
			err = sender.Send(tt.exchange, tt.key, body, map[string]interface{}{"Authorization": token})
			if err != nil {
				assert.FailNow(t, err.Error())
			}

			t.Log("subscribe to:", tt.respExchange, tt.respKey)
			chanBody, err := sender.Subscribe(tt.respExchange, tt.respKey, map[string]interface{}{})
			if err != nil {
				assert.FailNow(t, err.Error())
			}

			resp := tt.msgResp
			select {
			case receivedMsg := <-chanBody:
				err = json.Unmarshal(receivedMsg, &resp)
				if err != nil {
					assert.FailNow(t, err.Error())
				}
			case <-time.After(time.Second):
				assert.FailNow(t, "timeout waiting response")
			}

			switch tmp := resp.(type) {
			case network.DeviceRegisteredResponse:
				assert.Equal(t, nil, tmp.Error)
				assert.Equal(t, "123", tmp.ID)
				assert.Equal(t, "testThing", tmp.Name)
				assert.NotEmpty(t, tmp.Token)
			case network.SchemaUpdatedResponse:
				in, _ := tt.msg.(network.SchemaUpdateRequest)
				assert.Equal(t, nil, tmp.Error)
				assert.Equal(t, "123", tmp.ID)
				assert.Equal(t, in.Schema, tmp.Schema)
			case network.DataSent:
				assert.Equal(t, "123", tmp.ID)
			case network.DataRequest:
				assert.Equal(t, "123", tmp.ID)
			case network.DataUpdate:
				assert.Equal(t, "123", tmp.ID)
			case network.DeviceUnregisteredResponse:
				assert.Equal(t, nil, tmp.Error)
				assert.Equal(t, "123", tmp.ID)
			}
		})
	}
}

func registerThing(ID, name string) (string, error) {
	msg := network.DeviceRegisterRequest{ID: ID, Name: name}
	body, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	err = sender.Send("device", "device.register", body, map[string]interface{}{"Authorization": token})
	if err != nil {
		return "", err
	}

	chanBody, err := sender.Subscribe("device", "device.registered", map[string]interface{}{})
	if err != nil {
		return "", err
	}

	resp := network.DeviceRegisteredResponse{}
	select {
	case receivedMsg := <-chanBody:
		err = json.Unmarshal(receivedMsg, &resp)
		if err != nil {
			return "", err
		}

		return resp.Token, nil
	case <-time.After(time.Second):
		return "", errors.New("timeout waiting response")
	}
}

func unregisterThing(ID string) error {
	msg := network.DeviceUnregisterRequest{ID: ID}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = sender.Send("device", "device.unregister", body, map[string]interface{}{"Authorization": token})
	if err != nil {
		return err
	}

	chanBody, err := sender.Subscribe("device", "device.unregistered", map[string]interface{}{})
	if err != nil {
		return err
	}

	resp := network.DeviceUnregisteredResponse{}
	select {
	case receivedMsg := <-chanBody:
		err = json.Unmarshal(receivedMsg, &resp)
		if err != nil {
			return err
		}

		if resp.Error != nil {
			return errors.New(*resp.Error)
		}

		return nil
	case <-time.After(time.Second):
		return errors.New("timeout waiting response")
	}
}

func TestHappyPathRPCAuth(t *testing.T) {
	thingToken, err := registerThing("123", "testThing")
	if err != nil {
		t.FailNow()
	}
	defer func() {
		err = unregisterThing("123")
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}()
	tests := []struct {
		name  string
		ID    string
		token string
	}{
		{"with a thing registred, the happy path call auth should return no error", "123", thingToken},
		// TODO: add more tests
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := rpc.Auth(tt.ID, tt.token)

			assert.Nil(t, err)
			assert.Equal(t, tt.ID, id)
		})
	}
}
func TestHappyPathRPCList(t *testing.T) {
	thingToken, err := registerThing("123", "testThing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer func() {
		err = unregisterThing("123")
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}()

	tests := []struct {
		name  string
		ID    string
		token string
	}{
		{"with a thing registred, the happy path call list should return no error", "123", thingToken},
		// TODO: add more tests
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			things, err := rpc.List()

			assert.Nil(t, err)
			assert.Equal(t, 1, len(things))
			assert.Equal(t, "123", things[0].ID)
			assert.Equal(t, "testThing", things[0].Name)
			assert.Empty(t, things[0].Token)
			assert.Nil(t, things[0].Schema)
		})
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(0)
	}

	err := SetupSuite()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	code := m.Run()
	err = TearDownSuite()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	os.Exit(code)
}
