# Order Service - Go Study Project

Serviço de gerenciamento de pedidos desenvolvido em **Go** para estudo da linguagem, aplicando arquitetura limpa e event-driven architecture com RabbitMQ.

---

## Objetivo

Projeto de estudo focado em **Golang**, aplicando conhecimentos de:
- Arquitetura de microserviços
- Clean Architecture
- Mensageria assíncrona (RabbitMQ)
- APIs REST com Gin
- ORM (GORM) com PostgreSQL

---

## Stack

- **Go** 1.21+
- **Gin** - Framework HTTP
- **GORM** - ORM
- **PostgreSQL** - Database
- **RabbitMQ** - Message Broker
- **Docker** - Containerização

---

## Estrutura do Projeto

```
order-service/
├── cmd/
│   ├── order-service/              # Entrypoint principal
│   │   └── main.go
│   └── test-consumer/              # Consumer de teste
│       └── main.go
│
├── internal/
│   ├── config/
│   │   └── config.go               # Configurações (env vars)
│   │
│   └── order/
│       ├── handler/
│       │   └── order_handler.go    # HTTP Handlers (Controllers)
│       │
│       ├── model/
│       │   └── order.go            # Domain Models + DTOs
│       │
│       ├── repository/
│       │   └── order_repository.go # Data Access Layer
│       │
│       └── service/
│           └── order_service.go    # Business Logic
│
├── pkg/
│   ├── db/
│   │   └── db.go                   # Database Connection
│   │
│   ├── logger/
│   │   └── logger.go               # Logging Utils
│   │
│   └── mq/
│       ├── publisher.go            # RabbitMQ Publisher
│       └── consumer.go             # RabbitMQ Consumer
│
├── .env                            # Environment Variables
├── docker-compose.yml              # PostgreSQL + RabbitMQ
├── go.mod
├── go.sum
└── README.md
```

---

## Arquitetura em Camadas

```
HTTP Request
     ↓
┌─────────────────┐
│    Handler      │  ← Gin (routing, validation)
└────────┬────────┘
         ↓
┌─────────────────┐
│    Service      │  ← Business Logic + Events
└────────┬────────┘
         ↓
┌─────────────────┐
│   Repository    │  ← GORM (database access)
└────────┬────────┘
         ↓
┌─────────────────┐
│   PostgreSQL    │  ← Data Persistence
└─────────────────┘

         ↓ (async)
         
┌─────────────────┐
│    RabbitMQ     │  ← Event Publishing
└─────────────────┘
```

---

## API Endpoints

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `POST` | `/api/v1/orders` | Criar pedido |
| `GET` | `/api/v1/orders/:id` | Buscar pedido por ID |
| `GET` | `/api/v1/orders?customer_id=X` | Listar pedidos do cliente |
| `PUT` | `/api/v1/orders/:id/status` | Atualizar status |
| `PUT` | `/api/v1/orders/:id/cancel` | Cancelar pedido |
| `GET` | `/health` | Health check |

---

## Eventos RabbitMQ

**Exchange:** `orders_exchange` (tipo: topic)

**Eventos publicados:**
- `order.created` - Pedido criado
- `order.status_changed` - Status alterado
- `order.cancelled` - Pedido cancelado

---

## Como Executar

### 1. Clone e configure
```bash
git clone https://github.com/seu-usuario/order-service.git
cd order-service
```

### 2. Crie `.env`
```bash
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=orders_user
DB_PASSWORD=orders_pass
DB_NAME=orders_db
RABBITMQ_URL=amqp://orders_user:orders_pass@localhost:5672/
```

### 3. Instale dependências
```bash
go mod download
```

### 4. Suba infraestrutura
```bash
docker-compose up -d
```

### 5. Execute
```bash
go run cmd/order-service/main.go
```

---

## Cenários de Teste

### Preparação
```bash
# Terminal 1 - Order Service
go run cmd/order-service/main.go

# Terminal 2 - Consumer (para ver eventos)
go run cmd/test-consumer/main.go
```

### 1. Criar pedido
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "items": [
      {"product_id": 101, "name": "Notebook Dell", "price": 2599.99, "quantity": 1},
      {"product_id": 102, "name": "Mouse Logitech", "price": 89.90, "quantity": 2}
    ]
  }'
```

### 2. Buscar pedido por ID
```bash
curl http://localhost:8080/api/v1/orders/1
```

### 3. Listar pedidos do cliente
```bash
curl "http://localhost:8080/api/v1/orders?customer_id=1&limit=10&offset=0"
```

### 4. Atualizar status (fluxo completo)
```bash
# Pending → Confirmed
curl -X PUT http://localhost:8080/api/v1/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "confirmed"}'

# Confirmed → Paid
curl -X PUT http://localhost:8080/api/v1/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "paid"}'

# Paid → Shipped
curl -X PUT http://localhost:8080/api/v1/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}'

# Shipped → Delivered
curl -X PUT http://localhost:8080/api/v1/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "delivered"}'
```

### 5. Cancelar pedido
```bash
# Criar novo pedido
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 2,
    "items": [
      {"product_id": 999, "name": "Produto Teste", "price": 50.00, "quantity": 1}
    ]
  }'

# Cancelar
curl -X PUT http://localhost:8080/api/v1/orders/2/cancel
```

### 6. Health check
```bash
curl http://localhost:8080/health
```

### 7. Verificar RabbitMQ Management
**Acesse:** http://localhost:15672 (orders_user/orders_pass)

---

## Conceitos Go Aplicados

- ✅ Structs e Interfaces
- ✅ Error handling idiomático
- ✅ Goroutines e Channels
- ✅ Packages e Modules
- ✅ Dependency Injection
- ✅ Context

---

**Projeto desenvolvido para estudo de Go (Golang)**