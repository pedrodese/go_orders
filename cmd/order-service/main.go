package main

import (
	"log"
	"net/http"

	"order-service/internal/config"
	"order-service/internal/order/handler"
	"order-service/internal/order/repository"
	"order-service/internal/order/service"
	"order-service/pkg/db"
	"order-service/pkg/mq"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	log.Printf("🚀 Iniciando Order Service na porta %s", cfg.Server.Port)

	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}

	if err := db.Migrate(database); err != nil {
		log.Fatal("Erro ao executar migrations:", err)
	}

	publisher, err := mq.NewPublisher(&cfg.RabbitMQ)
	if err != nil {
		log.Fatal("Erro ao conectar ao RabbitMQ:", err)
	}
	defer publisher.Close()

	orderRepo := repository.NewOrderRepository(database)
	orderService := service.NewOrderService(orderRepo, publisher)
	orderHandler := handler.NewOrderHandler(orderService)

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "order-service",
			"version": "1.0.0",
		})
	})

	api := r.Group("/api/v1")
	{
		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.GetOrdersByCustomer)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.PUT("/:id/status", orderHandler.UpdateOrderStatus)
			orders.PUT("/:id/cancel", orderHandler.CancelOrder)
		}
	}

	log.Printf("Order Service rodando em http://localhost:%s", cfg.Server.Port)
	log.Printf("Health check: http://localhost:%s/health", cfg.Server.Port)
	log.Printf("API Orders: http://localhost:%s/api/v1/orders", cfg.Server.Port)

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Erro ao iniciar servidor:", err)
	}
}
