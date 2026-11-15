package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/JanKrl/homelab-version-agent/internal/checker"
	"github.com/JanKrl/homelab-version-agent/internal/config"
)

const waitTime uint = 250

type MqttPublisher struct {
	client     mqtt.Client
	topic      string
	host       string
	discovered map[string]bool
}

type mqttMessage struct {
	Host string `json:"host"`
	checker.Result
	LastCheckedAt string `json:"last_checked_at"`
}

type haDiscoveryDevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Model        string   `json:"model"`
	Manufacturer string   `json:"manufacturer"`
}

type haDiscoveryPayload struct {
	Name                  string            `json:"name"`
	UniqueID              string            `json:"unique_id"`
	StateTopic            string            `json:"state_topic"`
	ValueTemplate         string            `json:"value_template"`
	LatestVersionTopic    string            `json:"latest_version_topic"`
	LatestVersionTemplate string            `json:"latest_version_template"`
	ReleaseURL            string            `json:"release_url,omitempty"`
	AvailabilityTopic     string            `json:"availability_topic"`
	PayloadAvailable      string            `json:"payload_available"`
	PayloadNotAvailable   string            `json:"payload_not_available"`
	Device                haDiscoveryDevice `json:"device"`
}

func NewMqttPublisher(mqttConfig config.MqttConfig, host string) (*MqttPublisher, error) {
	client, err := initialize(mqttConfig, host)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MQTT client: %w", err)
	}

	return &MqttPublisher{
		client:     client,
		topic:      mqttConfig.Topic,
		host:       host,
		discovered: make(map[string]bool),
	}, nil
}

func (p *MqttPublisher) Publish(ctx context.Context, results []checker.Result) error {
	for _, result := range results {
		if !p.discovered[result.Name] {
			discTopic, discPayload, err := constructDiscoveryMessage(result, p.topic, p.host)
			if err != nil {
				return err
			}
			token := p.client.Publish(discTopic, 1, true, discPayload)
			if token.Wait() && token.Error() != nil {
				return fmt.Errorf("failed to publish discovery message to topic %s: %w", discTopic, token.Error())
			}
			p.discovered[result.Name] = true
		}

		topic := stateTopic(p.topic, p.host, result.Name)
		payloadBytes, err := constructMessage(result, p.host)
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

func (p *MqttPublisher) Close() {
	token := p.client.Publish(availabilityTopic(p.topic, p.host), 1, true, "offline")
	token.Wait()
	p.client.Disconnect(waitTime)
}

func constructMessage(result checker.Result, host string) ([]byte, error) {
	payload := mqttMessage{host, result, time.Now().Format(time.RFC3339)}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return payloadBytes, nil
}

func constructDiscoveryMessage(result checker.Result, baseTopic, host string) (string, []byte, error) {
	sanitizedHost := sanitizeID(host)
	sanitizedName := sanitizeID(result.Name)
	uniqueId := fmt.Sprintf("%s___%s", sanitizedHost, sanitizedName)

	discoveryTopic := discoveryTopic(sanitizedHost, sanitizedName)
	stateTopic := stateTopic(baseTopic, host, result.Name)

	device := haDiscoveryDevice{
		Identifiers:  []string{"homelab_lxc_" + sanitizedHost},
		Name:         "LXC " + host,
		Model:        "Proxmox LXC Container",
		Manufacturer: "Homelab",
	}

	discoveryPayload := haDiscoveryPayload{
		Name:                  result.Name,
		UniqueID:              uniqueId,
		StateTopic:            stateTopic,
		ValueTemplate:         "{{ value_json.current }}",
		LatestVersionTopic:    stateTopic,
		LatestVersionTemplate: "{{ value_json.latest }}",
		ReleaseURL:            result.ReleaseURL,
		AvailabilityTopic:     availabilityTopic(baseTopic, host),
		PayloadAvailable:      "online",
		PayloadNotAvailable:   "offline",
		Device:                device,
	}

	payloadBytes, err := json.Marshal(discoveryPayload)
	if err != nil {
		return "", nil, err
	}

	return discoveryTopic, payloadBytes, nil
}

func buildClientOptions(mqttConfig config.MqttConfig, host string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%v:%v", mqttConfig.Broker, mqttConfig.Port))
	opts.SetUsername(mqttConfig.Username)
	opts.SetPassword(mqttConfig.Password)
	opts.SetAutoReconnect(true)

	statusTopic := availabilityTopic(mqttConfig.Topic, host)
	opts.SetWill(statusTopic, "offline", 1, true)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		c.Publish(statusTopic, 1, true, "online")
	})

	return opts
}

func initialize(mqttConfig config.MqttConfig, host string) (mqtt.Client, error) {
	opts := buildClientOptions(mqttConfig, host)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

var idReplacer = strings.NewReplacer("/", "__", ".", "__", " ", "__")

func sanitizeID(s string) string {
	return idReplacer.Replace(s)
}

func availabilityTopic(baseTopic, host string) string {
	return fmt.Sprintf("%s/%s/status", baseTopic, host)
}

func discoveryTopic(nodeId, objectId string) string {
	return fmt.Sprintf("homeassistant/update/%s/%s/config", nodeId, objectId)
}

func stateTopic(baseTopic, host, name string) string {
	return fmt.Sprintf("%s/%s/%s", baseTopic, host, name)
}
