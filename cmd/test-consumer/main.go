package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"order-service/internal/config"
	"order-service/pkg/mq"
)

func main() {
	log.Println("Iniciando Test Consumer...")

	cfg := config.Load()

	consumer, err := mq.NewConsumer(&cfg.RabbitMQ)
	if err != nil {
		log.Fatal("Erro ao conectar consumer:", err)
	}
	defer consumer.Close()

	handler := func(event mq.OrderEvent) error {
		log.Printf("EVENTO RECEBIDO:")
		log.Printf("   Tipo: %s", event.Type)
		log.Printf("   Order ID: %d", event.OrderID)

		if jsonData, err := json.MarshalIndent(event.Data, "   ", "  "); err == nil {
			log.Printf("   Dados: %s", string(jsonData))
		}

		log.Println("   ---")
		return nil
	}

	routingKeys := []string{
		"order.created",
		"order.status_changed",
		"order.cancelled",
		"order.*", // Captura todos eventos de order
	}

	if err := consumer.StartListening("test_order_events", routingKeys, handler); err != nil {
		log.Fatal("Erro ao iniciar consumer:", err)
	}

	log.Println("Consumer rodando... Pressione Ctrl+C para parar")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Parando consumer...")
}
