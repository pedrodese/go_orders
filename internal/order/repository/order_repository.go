package repository

import (
	"order-service/internal/order/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *model.Order) error
	GetByID(id uint) (*model.Order, error)
	GetByCustomerID(customerID uint, limit, offset int) ([]model.Order, error)
	Update(order *model.Order) error
	UpdateStatus(id uint, status model.OrderStatus) error
	Delete(id uint) error
	Count() (int64, error)
	CountByCustomer(customerID uint) (int64, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order) error {
	// Calcular subtotais dos items
	for i := range order.Items {
		order.Items[i].CalculateSubtotal()
	}

	// Calcular total do pedido
	order.CalculateTotal()

	return r.db.Create(order).Error
}

func (r *orderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Items").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByCustomerID(customerID uint, limit, offset int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Preload("Items").
		Where("customer_id = ?", customerID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&orders).Error

	return orders, err
}

func (r *orderRepository) Update(order *model.Order) error {
	// Recalcular totais antes de salvar
	for i := range order.Items {
		order.Items[i].CalculateSubtotal()
	}
	order.CalculateTotal()

	return r.db.Save(order).Error
}

func (r *orderRepository) UpdateStatus(id uint, status model.OrderStatus) error {
	return r.db.Model(&model.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *orderRepository) Delete(id uint) error {
	return r.db.Delete(&model.Order{}, id).Error
}

func (r *orderRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Order{}).Count(&count).Error
	return count, err
}

func (r *orderRepository) CountByCustomer(customerID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Order{}).
		Where("customer_id = ?", customerID).
		Count(&count).Error
	return count, err
}
