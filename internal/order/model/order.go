package model

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusPaid      OrderStatus = "paid"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
	StatusFailed    OrderStatus = "failed"
)

type Order struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CustomerID  uint           `json:"customer_id" gorm:"not null"`
	Status      OrderStatus    `json:"status" gorm:"type:varchar(20);default:'pending'"`
	TotalAmount float64        `json:"total_amount" gorm:"type:decimal(10,2)"`
	Items       []OrderItem    `json:"items" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type OrderItem struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	OrderID   uint           `json:"order_id" gorm:"not null"`
	ProductID uint           `json:"product_id" gorm:"not null"`
	Name      string         `json:"name" gorm:"type:varchar(255);not null"`
	Price     float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	Quantity  int            `json:"quantity" gorm:"not null"`
	Subtotal  float64        `json:"subtotal" gorm:"type:decimal(10,2)"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateOrderRequest struct {
	CustomerID uint                     `json:"customer_id" binding:"required"`
	Items      []CreateOrderItemRequest `json:"items" binding:"required,dive"`
}

type CreateOrderItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price" binding:"required,min=0"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
}

type OrderResponse struct {
	ID          uint                `json:"id"`
	CustomerID  uint                `json:"customer_id"`
	Status      OrderStatus         `json:"status"`
	TotalAmount float64             `json:"total_amount"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ID        uint    `json:"id"`
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}

func (o *Order) CalculateTotal() {
	total := 0.0
	for _, item := range o.Items {
		total += item.Subtotal
	}
	o.TotalAmount = total
}

func (oi *OrderItem) CalculateSubtotal() {
	oi.Subtotal = oi.Price * float64(oi.Quantity)
}

func (o *Order) ToResponse() OrderResponse {
	items := make([]OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
			Subtotal:  item.Subtotal,
		}
	}

	return OrderResponse{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		TotalAmount: o.TotalAmount,
		Items:       items,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}
