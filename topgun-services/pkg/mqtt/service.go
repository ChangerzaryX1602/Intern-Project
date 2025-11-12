package mqtt

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Service struct {
	client mqtt.Client
	topic  string
}

func NewService(client mqtt.Client, topic string) *Service {
	return &Service{
		client: client,
		topic:  topic,
	}
}

// Subscribe subscribes to the topic with a custom message handler
func (s *Service) Subscribe(handler mqtt.MessageHandler) error {
	if handler == nil {
		handler = s.defaultMessageHandler
	}

	token := s.client.Subscribe(s.topic, 1, handler)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", s.topic, token.Error())
	}

	log.Printf("Subscribed to MQTT topic: %s", s.topic)
	return nil
}

// Publish publishes a message to the topic
func (s *Service) Publish(message string) error {
	if s.client == nil || !s.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	token := s.client.Publish(s.topic, 1, false, message)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to publish message: %w", token.Error())
	}

	log.Printf("Published message to %s: %s", s.topic, message)
	return nil
}

// PublishJSON publishes a JSON object to the topic
func (s *Service) PublishJSON(data interface{}) error {
	if s.client == nil || !s.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	token := s.client.Publish(s.topic, 1, false, jsonData)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to publish JSON message: %w", token.Error())
	}

	log.Printf("Published JSON message to %s", s.topic)
	return nil
}

// PublishBytes publishes raw byte data to the topic
func (s *Service) PublishBytes(data []byte) error {
	if s.client == nil || !s.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	token := s.client.Publish(s.topic, 1, false, data)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to publish bytes: %w", token.Error())
	}

	log.Printf("Published %d bytes to %s", len(data), s.topic)
	return nil
}

// PublishFileMetadata publishes file metadata as JSON (filename, size, chunks info)
func (s *Service) PublishFileMetadata(metadata map[string]interface{}) error {
	return s.PublishJSON(metadata)
}

// IsConnected checks if the client is connected
func (s *Service) IsConnected() bool {
	return s.client != nil && s.client.IsConnected()
}

// GetTopic returns the configured topic
func (s *Service) GetTopic() string {
	return s.topic
}

// Default message handler
func (s *Service) defaultMessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))
}
