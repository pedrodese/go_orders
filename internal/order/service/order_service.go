package service

import (
	"fmt"
	"log"
	"order-service/internal/order/model"
	"order-service/internal/order/repository"
	"order-service/pkg/mq"
	"slices"
	"time"
)

type OrderService interface {
	CreateOrder(req model.CreateOrderRequest) (*model.OrderResponse, error)
	GetOrderByID(id uint) (*model.OrderResponse, error)
	GetOrdersByCustomer(customerID uint, limit, offset int) ([]model.OrderResponse, error)
	UpdateOrderStatus(id uint, status model.OrderStatus) (*model.OrderResponse, error)
	CancelOrder(id uint) error
}

type orderService struct {
	orderRepo repository.OrderRepository
	publisher mq.Publisher
}

func NewOrderService(orderRepo repository.OrderRepository, publisher mq.Publisher) OrderService {
	return &orderService{
		orderRepo: orderRepo,
		publisher: publisher,
	}
}

func (s *orderService) CreateOrder(req model.CreateOrderRequest) (*model.OrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("pedido deve conter pelo menos um item")
	}

	order := &model.Order{
		CustomerID: req.CustomerID,
		Status:     model.StatusPending,
		Items:      make([]model.OrderItem, len(req.Items)),
	}

	for i, item := range req.Items {
		order.Items[i] = model.OrderItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
		}
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("erro ao criar pedido: %w", err)
	}

	log.Printf("Pedido criado: ID=%d, Customer=%d, Total=%.2f",
		order.ID, order.CustomerID, order.TotalAmount)

	if err := s.publishOrderCreatedEvent(order); err != nil {
		log.Printf("Erro ao publicar evento order.created: %v", err)
	}

	response := order.ToResponse()
	return &response, nil
}

func (s *orderService) GetOrderByID(id uint) (*model.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("pedido não encontrado: %w", err)
	}

	response := order.ToResponse()
	return &response, nil
}

func (s *orderService) GetOrdersByCustomer(customerID uint, limit, offset int) ([]model.OrderResponse, error) {
	orders, err := s.orderRepo.GetByCustomerID(customerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedidos do cliente: %w", err)
	}

	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}

	return responses, nil
}

func (s *orderService) UpdateOrderStatus(id uint, status model.OrderStatus) (*model.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("pedido não encontrado: %w", err)
	}

	if !s.isValidStatusTransition(order.Status, status) {
		return nil, fmt.Errorf("transição de status inválida: %s -> %s", order.Status, status)
	}

	if err := s.orderRepo.UpdateStatus(id, status); err != nil {
		return nil, fmt.Errorf("erro ao atualizar status: %w", err)
	}

	log.Printf("Status do pedido %d alterado: %s -> %s", id, order.Status, status)

	if err := s.publishOrderStatusChangedEvent(id, order.Status, status); err != nil {
		log.Printf("Erro ao publicar evento order.status_changed: %v", err)
	}

	return s.GetOrderByID(id)
}

func (s *orderService) CancelOrder(id uint) error {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("pedido não encontrado: %w", err)
	}

	if order.Status != model.StatusPending && order.Status != model.StatusConfirmed {
		return fmt.Errorf("não é possível cancelar pedido com status: %s", order.Status)
	}

	if err := s.orderRepo.UpdateStatus(id, model.StatusCancelled); err != nil {
		return fmt.Errorf("erro ao cancelar pedido: %w", err)
	}

	log.Printf("Pedido %d cancelado", id)

	if err := s.publishOrderCancelledEvent(id); err != nil {
		log.Printf("Erro ao publicar evento order.cancelled: %v", err)
	}

	return nil
}

func (s *orderService) isValidStatusTransition(from, to model.OrderStatus) bool {
	validTransitions := map[model.OrderStatus][]model.OrderStatus{
		model.StatusPending: {
			model.StatusConfirmed,
			model.StatusCancelled,
		},
		model.StatusConfirmed: {
			model.StatusPaid,
			model.StatusCancelled,
		},
		model.StatusPaid: {
			model.StatusShipped,
		},
		model.StatusShipped: {
			model.StatusDelivered,
		},
		model.StatusDelivered: {},
		model.StatusCancelled: {},
		model.StatusFailed:    {},
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}

	return slices.Contains(allowedTransitions, to)
}

func (s *orderService) publishOrderCreatedEvent(order *model.Order) error {
	eventData := map[string]any{
		"order_id":     order.ID,
		"customer_id":  order.CustomerID,
		"status":       order.Status,
		"total_amount": order.TotalAmount,
		"items":        order.Items,
		"created_at":   order.CreatedAt,
	}

	return s.publisher.PublishOrderEvent("created", order.ID, eventData)
}

func (s *orderService) publishOrderStatusChangedEvent(orderID uint, oldStatus, newStatus model.OrderStatus) error {
	eventData := map[string]any{
		"order_id":   orderID,
		"old_status": oldStatus,
		"new_status": newStatus,
		"changed_at": time.Now(),
	}

	return s.publisher.PublishOrderEvent("status_changed", orderID, eventData)
}

func (s *orderService) publishOrderCancelledEvent(orderID uint) error {
	eventData := map[string]any{
		"order_id":     orderID,
		"cancelled_at": time.Now(),
	}

	return s.publisher.PublishOrderEvent("cancelled", orderID, eventData)
}
