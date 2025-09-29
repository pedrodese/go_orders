package mq

import (
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer interface {
	StartListening(queueName string, routingKeys []string, handler EventHandler) error
	Close() error
}

type EventHandler func(event OrderEvent) error

type consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *config.RabbitMQConfig
}

func NewConsumer(cfg *config.RabbitMQConfig) (Consumer, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	c := &consumer{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}

	// Declarar exchange (caso não exista)
	if err := c.setupExchange(); err != nil {
		c.Close()
		return nil, err
	}

	log.Printf("RabbitMQ Consumer conectado")
	return c, nil
}

func (c *consumer) setupExchange() error {
	return c.channel.ExchangeDeclare(
		c.config.Exchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
}

func (c *consumer) StartListening(queueName string, routingKeys []string, handler EventHandler) error {
	// Declarar fila
	queue, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind fila ao exchange com routing keys
	for _, routingKey := range routingKeys {
		err := c.channel.QueueBind(
			queue.Name,        // queue name
			routingKey,        // routing key
			c.config.Exchange, // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
		log.Printf("Fila '%s' vinculada ao routing key: %s", queueName, routingKey)
	}

	// Configurar QoS
	err = c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Consumir mensagens
	msgs, err := c.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Escutando mensagens na fila: %s", queueName)

	// Processar mensagens em goroutine
	go func() {
		for msg := range msgs {
			if err := c.processMessage(msg, handler); err != nil {
				log.Printf("Erro ao processar mensagem: %v", err)
				msg.Nack(false, true) // Rejeitar e reenviar para fila
			} else {
				msg.Ack(false)
			}
		}
	}()

	return nil
}

func (c *consumer) processMessage(msg amqp.Delivery, handler EventHandler) error {
	var event OrderEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	log.Printf("Evento recebido: %s (Order ID: %d)", event.Type, event.OrderID)

	return handler(event)
}

func (c *consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
