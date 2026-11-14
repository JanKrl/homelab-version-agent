package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/JanKrl/homelab-version-agent/internal/checker"
	"github.com/JanKrl/homelab-version-agent/internal/config"
)

const waitTime uint = 250

type MqttPublisher struct {
	client mqtt.Client
	topic  string
}

type mqttMessage struct {
	checker.Result
	LastCheckedAt string `json:"last_checked_at"`
}

func NewMqttPublisher(mqttConfig config.MqttConfig) (*MqttPublisher, error) {
	client, err := initialize(mqttConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MQTT client: %w", err)
	}

	return &MqttPublisher{
		client: client,
		topic:  mqttConfig.Topic,
	}, nil
}

func (p *MqttPublisher) Publish(ctx context.Context, results []checker.Result) error {
	token := p.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	defer p.client.Disconnect(waitTime)

	for _, result := range results {
		topic := fmt.Sprintf("%s/%s", p.topic, result.Name)
		payloadBytes, err := constructMessage(result)
		if err != nil {
			return err
		}
		// QoS = 1 makes sure that package is received before disconnecting from broker
		// Retained = true makes sure that broker saves last published state
		token := p.client.Publish(topic, 1, true, payloadBytes)
		if token.Wait() && token.Error() != nil {
			return fmt.Errorf("failed to publish message to topic %s: %w", topic, token.Error())
		}
	}

	return nil
}

func constructMessage(result checker.Result) ([]byte, error) {
	payload := mqttMessage{result, time.Now().Format(time.RFC3339)}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return payloadBytes, nil
}

func initialize(mqttConfig config.MqttConfig) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%v:%v", mqttConfig.Broker, mqttConfig.Port))
	opts.SetUsername(mqttConfig.Username)
	opts.SetPassword(mqttConfig.Password)

	client := mqtt.NewClient(opts)
	err := checkConnection(client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func checkConnection(client mqtt.Client) error {
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	defer client.Disconnect(waitTime)

	return nil
}
