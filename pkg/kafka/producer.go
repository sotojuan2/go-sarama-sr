package kafka

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/go-sarama-sr/producer/internal/config"
)

// Producer wraps the Sarama sync producer with our configuration
type Producer struct {
	producer sarama.SyncProducer
	config   *config.Config
}

// NewProducer creates a new Kafka producer with Confluent Cloud configuration
func NewProducer(cfg *config.Config) (*Producer, error) {
	// Configure Sarama for Confluent Cloud with SASL_SSL
	saramaConfig := sarama.NewConfig()
	
	// Security configuration for Confluent Cloud
	saramaConfig.Net.SASL.Enable = true
	saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	saramaConfig.Net.SASL.User = cfg.Kafka.APIKey
	saramaConfig.Net.SASL.Password = cfg.Kafka.APISecret
	
	// TLS configuration
	saramaConfig.Net.TLS.Enable = true
	saramaConfig.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: false,
	}
	
	// Producer configuration
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll  // Wait for all replicas
	saramaConfig.Producer.Retry.Max = 3                     // Retry up to 3 times
	saramaConfig.Producer.Return.Successes = true           // Return success messages
	
	// Create the producer
	producer, err := sarama.NewSyncProducer([]string{cfg.Kafka.BootstrapServers}, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Printf("✅ Successfully connected to Kafka cluster: %s", cfg.Kafka.BootstrapServers)
	
	return &Producer{
		producer: producer,
		config:   cfg,
	}, nil
}

// SendMessage sends a simple string message to the configured topic
func (p *Producer) SendMessage(message string) error {
	// Create the message
	msg := &sarama.ProducerMessage{
		Topic: p.config.Kafka.Topic,
		Value: sarama.StringEncoder(message),
	}

	// Send the message
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("✅ Message sent successfully to topic '%s' [partition=%d, offset=%d]: %s", 
		p.config.Kafka.Topic, partition, offset, message)
	
	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	if p.producer != nil {
		err := p.producer.Close()
		if err != nil {
			return fmt.Errorf("failed to close producer: %w", err)
		}
		log.Println("✅ Producer closed successfully")
	}
	return nil
}

// ListTopics lists all available topics in the cluster
func ListTopics(cfg *config.Config) ([]string, error) {
	// Configure Sarama for Confluent Cloud with SASL_SSL
	saramaConfig := sarama.NewConfig()
	
	// Security configuration for Confluent Cloud
	saramaConfig.Net.SASL.Enable = true
	saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	saramaConfig.Net.SASL.User = cfg.Kafka.APIKey
	saramaConfig.Net.SASL.Password = cfg.Kafka.APISecret
	
	// TLS configuration
	saramaConfig.Net.TLS.Enable = true
	saramaConfig.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: false,
	}
	
	// Create a client to list topics
	client, err := sarama.NewClient([]string{cfg.Kafka.BootstrapServers}, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}
	defer client.Close()

	// Get topic metadata
	topics, err := client.Topics()
	if err != nil {
		return nil, fmt.Errorf("failed to get topics: %w", err)
	}

	return topics, nil
}

// TopicExists checks if a specific topic exists in the cluster
func TopicExists(cfg *config.Config, topicName string) (bool, error) {
	topics, err := ListTopics(cfg)
	if err != nil {
		return false, err
	}

	for _, topic := range topics {
		if topic == topicName {
			return true, nil
		}
	}
	return false, nil
}
