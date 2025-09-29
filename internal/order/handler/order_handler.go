package handler

import (
	"net/http"
	"strconv"

	"order-service/internal/order/model"
	"order-service/internal/order/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req model.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	order, err := h.orderService.CreateOrder(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao criar pedido",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "ID deve ser um número",
		})
		return
	}

	order, err := h.orderService.GetOrderByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Pedido não encontrado",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetOrdersByCustomer(c *gin.Context) {
	customerIDStr := c.Query("customer_id")
	if customerIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Parâmetro obrigatório",
			Message: "customer_id é obrigatório",
		})
		return
	}

	customerID, err := strconv.ParseUint(customerIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "customer_id inválido",
			Message: "customer_id deve ser um número",
		})
		return
	}

	// Parâmetros de paginação
	limit := 10
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	orders, err := h.orderService.GetOrdersByCustomer(uint(customerID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Erro ao buscar pedidos",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, OrderListResponse{
		Orders: orders,
		Count:  len(orders),
		Limit:  limit,
		Offset: offset,
	})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "ID deve ser um número",
		})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Dados inválidos",
			Message: err.Error(),
		})
		return
	}

	order, err := h.orderService.UpdateOrderStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Erro ao atualizar status",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "ID inválido",
			Message: "ID deve ser um número",
		})
		return
	}

	if err := h.orderService.CancelOrder(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Erro ao cancelar pedido",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type OrderListResponse struct {
	Orders []model.OrderResponse `json:"orders"`
	Count  int                   `json:"count"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

type UpdateStatusRequest struct {
	Status model.OrderStatus `json:"status" binding:"required"`
}
