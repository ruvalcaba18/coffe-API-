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
The API features an **Automatic Migration Runner**. Just start the application and it will automatically synchronize the schema:
1. `01_init.sql` (Schema & Seed)
2. `02_reviews_favorites.sql`
3. `03_add_user_language.sql`
4. `04_shopping_cart.sql`
5. `05_add_user_avatar.sql`
6. `06_add_user_roles.sql`
7. `07_add_coupons.sql`

### 4. Running the App
```bash
go mod tidy
go run main.go
```

## 📡 RESTful API Endpoints

### 🔐 Session & Users
- `POST /api/v1/users`: Register a new account.
- `POST /api/v1/tokens`: Login and receive JWT access token.

### 👤 Profile
- `GET /api/v1/profile`: View current profile.
- `PATCH /api/v1/profile`: Update profile settings (e.g., language).
- `POST /api/v1/profile/avatar`: Upload profile picture (max 2MB, JPG/PNG).

### ☕️ Products
- `GET /api/v1/products`: List products with advanced filtering.
    - `q`: Search by name or description (ILIKE).
    - `category`: Filter by specific category.
    - `min_price`: Minimum price filter.
    - `max_price`: Maximum price filter.
- `GET /api/v1/products/{id}`: View product details (Redis Cached).
- `GET /api/v1/reviews?product_id={id}`: View product reviews.
- `POST /api/v1/reviews`: Submit a review (product_id in body).

### 🛒 Shopping Cart
- `GET /api/v1/cart`: View current cart.
- `PATCH /api/v1/cart`: Update item quantity.

### 📦 Orders
- `POST /api/v1/orders`: Create an order from cart (**ACID** Transaction). Supports optional `coupon_code` in JSON body.
- `GET /api/v1/orders`: View order history.

### ❤️ Favorites
- `GET /api/v1/favorites`: List your favorite products.
- `POST /api/v1/favorites`: Add to favorites (product_id in body).
- `DELETE /api/v1/favorites/{id}`: Remove from favorites.

### 🔔 Notifications
- `GET /api/v1/notifications/ws`: WebSocket connection for real-time status updates.

### 🛡 Admin Dashboard (Admin Role Required)
- `POST /api/v1/admin/products`: Create new products.
- `PUT /api/v1/admin/products/{id}`: Update product details.
- `DELETE /api/v1/admin/products/{id}`: Delete a product.
- `GET /api/v1/admin/orders`: View all system orders.
- `PATCH /api/v1/admin/orders/{id}`: Update order status (Triggers real-time WS notification).
- `POST /api/v1/admin/coupons`: Create a new discount coupon (supports fixed/percentage, expiry dates, and usage limits).
- `GET /api/v1/admin/coupons`: List all coupons and their usage status.
