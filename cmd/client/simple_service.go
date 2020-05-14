package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/CESARBR/knot-babeltower/internal/config"
	"github.com/CESARBR/knot-babeltower/pkg/network"
	"github.com/CESARBR/knot-babeltower/pkg/thing/entities"
)

type RPCService interface {
	Auth(string, string) (string, error)
	List() ([]*entities.Thing, error)
}

type simpleService struct {
	client    SimpleClient
	authToken string
}

func NewSimpleService(config config.RabbitMQ, token string) RPCService {
	client := NewSimpleSender()
	_ = client.Connect(config)
	return &simpleService{client: client, authToken: token}
}

func (s *simpleService) Auth(id string, token string) (string, error) {
	channel, err := s.client.Subscribe("device", "reply", nil)
	if err != nil {
		return "", err
	}
	fmt.Println("queue name: ", s.client.GetQueueName())
	req := network.DeviceAuthRequest{ID: id, Token: token}
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	err = s.client.Send("device", "device.auth", body, map[string]interface{}{"Authorization": s.authToken, "correlation_id": "1", "reply_to": "reply"})
	if err != nil {
		return "", err
	}

	select {
	case resp := <-channel:
		msg := network.DeviceAuthResponse{}
		if err := json.Unmarshal(resp, &msg); err != nil {
			return "", err
		}

		if msg.Error != nil {
			return "", errors.New(*msg.Error)
		}

		return msg.ID, nil
	case <-time.After(time.Second):
		return "", errors.New("timeout waiting response")
	}
}

func (s *simpleService) List() ([]*entities.Thing, error) {
	channel, err := s.client.Subscribe("device", "reply", nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("queue name: ", s.client.GetQueueName())

	err = s.client.Send("device", "device.list", []byte("{}"), map[string]interface{}{"Authorization": s.authToken, "correlation_id": "1", "reply_to": "reply"})
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-channel:
		msg := network.DeviceListResponse{}
		if err := json.Unmarshal(resp, &msg); err != nil {
			return nil, err
		}

		if msg.Error != nil {
			return nil, errors.New(*msg.Error)
		}

		return msg.Things, nil
	case <-time.After(time.Second):
		return nil, errors.New("timeout waiting response")
	}
}
