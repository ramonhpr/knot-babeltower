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

// GetTestConfig local configuration default
func GetTestConfig() config.Config {
	return config.Config{
		Server:   config.Server{Port: 8080},
		Logger:   config.Logger{Level: "debug"},
		Users:    config.Users{Hostname: "users", Port: 8180},
		RabbitMQ: config.RabbitMQ{URL: "amqp://rabbitmq"},
		Things:   config.Things{Hostname: "things", Port: 8180},
	}
}

func getToken(config config.Users) (string, error) {
	proxy := http.NewUserProxy(&mocks.FakeLogger{}, config.Hostname, config.Port)
	user := entities.User{Email: "test@test.com", Password: "12345678"}
	_ = proxy.Create(user)

	return proxy.CreateToken(user)
}

func SetupSuite() error {
	quitMain = make(chan bool, 1)
	started := make(chan bool, 1)
	config := GetTestConfig()
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

// Two clients send and each one receives the response
// One publisher and two subscribers
// Two publishers and one subscriber should receive both messages
// Test invalid configuration should return error/panic

func TestPubSub(t *testing.T) { // Actually should be RPC too
	tests := []struct {
		name         string
		exchange     string
		key          string
		respExchange string
		respKey      string
		token        string
		msg          interface{}
		msgResp      interface{}
		expectedErr  bool
	}{
		{
			"Happy path client send register then receive response",
			"device",
			"device.register",
			"device",
			"device.registered",
			token,
			network.DeviceRegisterRequest{ID: "123", Name: "testThing"},
			network.DeviceRegisteredResponse{},
			false,
		},
		{
			"Happy path client send schema then receive response",
			"device",
			"device.schema.sent",
			"device",
			"device.schema.updated",
			token,
			network.SchemaUpdateRequest{
				ID: "123",
				Schema: []thingEntities.Schema{
					{SensorID: 1, ValueType: 2, Unit: 1, TypeID: 13, Name: "testSensor"},
				},
			},
			network.SchemaUpdatedResponse{},
			false,
		},
		{
			"If client send register with already registered thing test should receive error",
			"device",
			"device.register",
			"device",
			"device.registered",
			token,
			network.DeviceRegisterRequest{ID: "123", Name: "testThing"},
			network.DeviceRegisteredResponse{},
			true,
		},
		{
			"Happy path client send unregister then receive response",
			"device",
			"device.unregister",
			"device",
			"device.unregistered",
			token,
			network.DeviceUnregisterRequest{ID: "123"},
			network.DeviceUnregisteredResponse{},
			false,
		},
		{
			"If client register Invalid ID should receive error",
			"device",
			"device.register",
			"device",
			"device.registered",
			token,
			network.DeviceRegisterRequest{ID: "invalid ID", Name: "testThing"},
			network.DeviceRegisteredResponse{},
			true,
		},
		{
			"If client register too long ID should receive error",
			"device",
			"device.register",
			"device",
			"device.registered",
			token,
			network.DeviceRegisterRequest{ID: "01234567890123456789", Name: "testThing"},
			network.DeviceRegisteredResponse{},
			true,
		},
		{
			"If client register empty ID should receive error",
			"device",
			"device.register",
			"device",
			"device.registered",
			token,
			network.DeviceRegisterRequest{ID: "", Name: "testThing"},
			network.DeviceRegisteredResponse{},
			true,
		},
		{
			"If client register empty Name should receive error",
			"device",
			"device.register",
			"device",
			"device.registered",
			token,
			network.DeviceRegisterRequest{ID: "123", Name: ""},
			network.DeviceRegisteredResponse{},
			true,
		},
		{
			"If client sends schema with invalid ID then receive error",
			"device",
			"device.schema.sent",
			"device",
			"device.schema.updated",
			token,
			network.SchemaUpdateRequest{
				ID: "invalid ID",
				Schema: []thingEntities.Schema{
					{SensorID: 1, ValueType: 2, Unit: 1, TypeID: 13, Name: "testSensor"},
				},
			},
			network.SchemaUpdatedResponse{},
			true,
		},
		{
			"If client sends unregister with Invalid ID then receive error",
			"device",
			"device.unregister",
			"device",
			"device.unregistered",
			token,
			network.DeviceUnregisterRequest{ID: "invalid id"},
			network.DeviceUnregisteredResponse{},
			true,
		},
	}

	for _, tt := range tests {
		assertFunc := func(t *testing.T) {
			err := subcribeAndSend(tt.msg, tt.exchange, tt.key, tt.token, &tt.msgResp, tt.respExchange, tt.respKey)
			if err != nil {
				assert.FailNow(t, err.Error())
			}

			switch tmp := tt.msgResp.(type) {
			case network.DeviceRegisteredResponse:
				assert.Equal(t, "123", tmp.ID)
				assert.Equal(t, "testThing", tmp.Name)
				if assert.NotNil(t, tmp.Error) {
					t.Log(*tmp.Error)
					assert.True(t, tt.expectedErr)
					assert.Empty(t, tmp.Token)
				} else {
					assert.False(t, tt.expectedErr)
					assert.NotEmpty(t, tmp.Token)
				}
			case network.SchemaUpdatedResponse:
				in, _ := tt.msg.(network.SchemaUpdateRequest)
				assert.Equal(t, "123", tmp.ID)
				assert.Equal(t, in.Schema, tmp.Schema)
				if assert.NotNil(t, tmp.Error) {
					t.Log(*tmp.Error)
					assert.True(t, tt.expectedErr)
				} else {
					assert.False(t, tt.expectedErr)
				}
			case network.DeviceUnregisteredResponse:
				assert.Equal(t, "123", tmp.ID)
				if assert.NotNil(t, tmp.Error) {
					t.Log(*tmp.Error)
					assert.True(t, tt.expectedErr)
				} else {
					assert.False(t, tt.expectedErr)
				}
			}
		}
		t.Run(tt.name, assertFunc)
		tt.expectedErr = true
		tt.token = ""
		t.Run("MissingToken/"+tt.name, assertFunc)
	}
}

func TestDataEvents(t *testing.T) {
	_, err := registerThing("123", "testThing")
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
		name         string
		exchange     string
		key          string
		respExchange string
		respKey      string
		token        string
		msg          interface{}
		msgResp      interface{}
		expectedErr  bool
	}{

		{
			"Happy path client send data then receive event published",
			"data.sent",
			"",
			"data.published",
			"",
			token,
			network.DataSent{
				ID: "123",
				Data: []thingEntities.Data{
					{SensorID: 1, Value: 12.5},
				},
			},
			network.DataSent{},
			false,
		},
		{
			"Happy path client send request data then receive response",
			"device",
			"data.request",
			"device",
			"device.123.data.request",
			token,
			network.DataRequest{
				ID:        "123",
				SensorIds: []int{1},
			},
			network.DataRequest{},
			false,
		},
		{
			"Happy path client send update data then receive response",
			"device",
			"data.update",
			"device",
			"device.123.data.update",
			token,
			network.DataUpdate{
				ID: "123",
				Data: []thingEntities.Data{
					{SensorID: 1, Value: 20.7},
				}},
			network.DataUpdate{},
			false,
		},
		// Can not use data before call schema
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := subcribeAndSend(tt.msg, tt.exchange, tt.key, tt.token, &tt.msgResp, tt.respExchange, tt.respKey)
			if err != nil {
				assert.FailNow(t, err.Error())
			}

			switch tmp := tt.msgResp.(type) {
			case network.DataSent:
				assert.Equal(t, "123", tmp.ID)
			case network.DataRequest:
				assert.Equal(t, "123", tmp.ID)
			case network.DataUpdate:
				assert.Equal(t, "123", tmp.ID)
			}
		})
	}
}

func subcribeAndSend(msgIn interface{}, exchangeIn, keyIn, authToken string, msgOut *interface{}, exchangeOut, keyOut string) error {
	body, err := json.Marshal(msgIn)
	if err != nil {
		return err
	}

	chanBody, err := sender.Subscribe(exchangeOut, keyOut, map[string]interface{}{})
	if err != nil {
		return err
	}

	err = sender.Send(exchangeIn, keyIn, body, map[string]interface{}{"Authorization": authToken})

	if err != nil {
		return err
	}

	select {
	case receivedMsg := <-chanBody:
		return json.Unmarshal(receivedMsg, msgOut)
	case <-time.After(time.Second):
		return errors.New("timeout waiting response")
	}
}

func registerThing(ID, name string) (string, error) {
	var resp interface{} = network.DeviceRegisteredResponse{}
	err := subcribeAndSend(network.DeviceRegisterRequest{ID: ID, Name: name}, "device", "device.register", token, &resp, "device", "device.registered")
	if err != nil {
		return "", err
	}

	tmp, ok := resp.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("incompatible type %T", resp)
	}

	if tmp["error"] != nil {
		errMsg, ok := tmp["error"].(*string)
		if !ok {
			return "", fmt.Errorf("incompatible type %T", tmp["error"])
		}

		if *errMsg != "" {
			return "", fmt.Errorf(*errMsg)
		}
	}

	return tmp["token"].(string), nil
}

func unregisterThing(ID string) error {
	var resp interface{} = &network.DeviceRegisteredResponse{}
	return subcribeAndSend(network.DeviceUnregisterRequest{ID: ID}, "device", "device.unregister", token, &resp, "device", "device.unregistered")
}

func TestHappyPathRPCAuth(t *testing.T) {
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
	_, err := registerThing("123", "testThing")
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
		name string
	}{
		{"with a thing registred, the happy path call list should return no error"},
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
