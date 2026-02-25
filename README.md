# CoffeeBase API ☕️

A professional, high-performance Coffee Shop API built with **Go**, **PostgreSQL**, and **Redis**. Designed to scale up to 50,000+ users with strict **ACID** compliance and advanced security features.

## 🚀 Key Features

- **Decoupled Architecture**: Highly modular structure with separate packages for models, stores, handlers, and services.
- **ACID Transactions**: Critical business logic (like Checkout) is handled in a Service layer using SQL transactions to ensure data integrity.
- **High-Performance Caching**: Redis integration for Products and Shopping Carts using the Cache-Aside pattern.
- **Advanced Security**: 
    - **JWT Authentication**: Secure endpoints with a 2-hour token expiration.
    - **Global Rate Limiting**: Redis-based middleware to prevent abuse (default: 60 req/min per IP).
    - **SQL Injection Prevention**: Full use of parameterized queries.
- **Idempotency**: Distributed locking in Redis to prevent duplicate orders during checkout.
- **Multilingual Support**: User-specific language preferences (es, en, fr, de, gsw).

## 📁 Project Structure

```text
.
├── main.go                 # Application entry point & dependency injection
├── api/
│   ├── handlers/           # HTTP controllers separated by domain
│   └── routes/             # Router configuration & middleware chain
├── internal/
│   ├── auth/               # JWT and Password utilities
│   ├── database/           # Postgres & Redis connection helpers
│   ├── middleware/         # Auth and Rate Limit middlewares
│   ├── models/             # Domain entities (User, Product, Order, Cart, etc.)
│   ├── service/            # Business logic & ACID transaction orchestration
│   └── store/              # Data persistence layer (PostgreSQL)
└── README.md
```

## 🛠 Tech Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL (Data Integrity)
- **Cache**: Redis (Speed & Locking)
- **Router**: Go-Chi (Lightweight & Fast)
- **Security**: JWT, Bcrypt, Redis-Rate-Limit

## 🚦 Getting Started

### 1. Requirements
- Go 1.24+
- PostgreSQL
- Redis

### 2. Environment Variables
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=yourpassword
export DB_NAME=coffeeshop
export REDIS_ADDR=localhost:6379
export JWT_SECRET=your_super_secret_key
```

### 3. Database Setup
Execute the following migrations in order from `internal/database/migrations/`:
1. `01_init.sql` (Schema & Seed)
2. `02_reviews_favorites.sql`
3. `03_add_user_language.sql`
4. `04_shopping_cart.sql`

### 4. Running the App
```bash
go mod tidy
go run main.go
```

## 📡 API Endpoints

### Authentication (Public)
- `POST /api/v1/auth/register`: Create user with language preference.
- `POST /api/v1/auth/login`: Get JWT (2h TTL).

### Products (Auth Required)
- `GET /api/v1/products`: List all (Cached in Redis).
- `GET /api/v1/products/{id}`: View details (Cached in Redis).
- `POST /api/v1/products/{id}/reviews`: Submit a review.
- `GET /api/v1/products/{id}/reviews`: View product reviews.
- `POST /api/v1/products/{id}/favorite`: Add to favorites.

### Shopping Cart (Auth Required)
- `GET /api/v1/cart`: View current cart (Fast Redis lookup).
- `PUT /api/v1/cart`: Add/Update item quantity.

### Orders (Auth Required)
- `POST /api/v1/orders/checkout`: **ACID** process to purchase cart items. Includes idempotency lock.
- `GET /api/v1/orders/history`: View past purchases.

### User Profile
- `GET /api/v1/me`: View current user data.
- `PUT /api/v1/me/language`: Update UI language preference.
