package mqtt

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client mqtt.Client
	topic  string
}

// NewMQTTClient creates a new MQTT client and connects to the broker
func NewMQTTClient(broker, clientID, topic string) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(messageHandler)
	opts.SetConnectionLostHandler(connectionLostHandler)
	opts.SetOnConnectHandler(onConnectHandler)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetMaxReconnectInterval(30 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	log.Printf("Connected to MQTT broker: %s", broker)

	mqttClient := &MQTTClient{
		client: client,
		topic:  topic,
	}

	return mqttClient, nil
}

// Subscribe subscribes to the configured topic with a custom message handler
func (m *MQTTClient) Subscribe(handler mqtt.MessageHandler) error {
	if handler == nil {
		handler = messageHandler
	}

	token := m.client.Subscribe(m.topic, 1, handler)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", m.topic, token.Error())
	}

	log.Printf("Subscribed to topic: %s", m.topic)
	return nil
}

// Publish publishes a message to the configured topic
func (m *MQTTClient) Publish(message string) error {
	token := m.client.Publish(m.topic, 1, false, message)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to publish message: %w", token.Error())
	}

	log.Printf("Published message to %s: %s", m.topic, message)
	return nil
}

// PublishJSON publishes a JSON message to the configured topic
func (m *MQTTClient) PublishJSON(jsonData []byte) error {
	token := m.client.Publish(m.topic, 1, false, jsonData)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to publish JSON message: %w", token.Error())
	}

	log.Printf("Published JSON message to %s", m.topic)
	return nil
}

// Disconnect disconnects from the MQTT broker
func (m *MQTTClient) Disconnect() {
	m.client.Disconnect(250)
	log.Println("Disconnected from MQTT broker")
}

// IsConnected checks if the client is connected to the broker
func (m *MQTTClient) IsConnected() bool {
	return m.client.IsConnected()
}

// Default message handler
var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message on topic %s: %s", msg.Topic(), msg.Payload())
}

// Connection lost handler
var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection lost: %v", err)
}

// On connect handler
var onConnectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to MQTT broker")
}
