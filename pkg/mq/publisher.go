package mq

import (
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	PublishOrderEvent(eventType string, orderID uint, data any) error
	Close() error
}

type publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *config.RabbitMQConfig
}

func NewPublisher(cfg *config.RabbitMQConfig) (Publisher, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	p := &publisher{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}

	// Declarar exchange
	if err := p.setupExchange(); err != nil {
		p.Close()
		return nil, err
	}

	log.Printf("RabbitMQ Publisher conectado")
	return p, nil
}

func (p *publisher) setupExchange() error {
	return p.channel.ExchangeDeclare(
		p.config.Exchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
}

func (p *publisher) PublishOrderEvent(eventType string, orderID uint, data interface{}) error {
	// Criar evento
	event := OrderEvent{
		Type:    eventType,
		OrderID: int(orderID),
		Data:    data,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	routingKey := fmt.Sprintf("order.%s", eventType)

	err = p.channel.Publish(
		p.config.Exchange, // exchange
		routingKey,        // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Persistir mensagem
			Body:         body,
			Headers: amqp.Table{
				"event_type": eventType,
				"order_id":   int(orderID),
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Evento publicado: %s (Order ID: %d)", eventType, orderID)
	return nil
}

func (p *publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	return nil
}

type OrderEvent struct {
	Type    string `json:"type"`
	OrderID int    `json:"order_id"`
	Data    any    `json:"data"`
}
