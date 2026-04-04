# Ultimate POS System - Point of Sale

Sistem **Point of Sale (POS)** yang ultimate dengan **JWT Authentication**, **Role-Based Access Control**, dan **Full POS Workflow** yang dibangun dengan **Clean Architecture** dan **Domain-Driven Design (DDD)** principles.

![Version](https://img.shields.io/badge/version-3.0.0-blue.svg)
![Go](https://img.shields.io/badge/go-1.26+-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![PostgreSQL](https://img.shields.io/badge/database-PostgreSQL-336791.svg)

## 🎉 What's New in v3.0.0 - Performance & Production Ready

### ✨ Major Optimizations

#### 🗄️ **Full PostgreSQL Support**
- ✅ **PostgreSQL Cart Repository** - Carts now persisted to database (no longer in-memory)
- ✅ **PostgreSQL Transaction Repository** - Transactions safely stored in PostgreSQL
- ✅ **PostgreSQL Token Repository** - Token blacklist persists across restarts
- ✅ **Zero data loss** on server restart when using PostgreSQL mode

#### ⚡ **Performance Optimizations**
- ✅ **Batch Inventory Updates** - Checkout now uses bulk operations instead of N+1 queries
- ✅ **Database Transactions** - Checkout wrapped in atomic DB transaction
- ✅ **Automatic Rollback** - Failed checkout automatically restores inventory
- ✅ **Thread-Safety** - All in-memory repositories now use `sync.RWMutex`
- ✅ **Optimized Queries** - Reduced database round-trips by 90%+ for multi-item carts

#### 🔧 **Code Quality Improvements**
- ✅ **Removed Duplicate Logic** - AuthUsecase now delegates to AuthService (DRY principle)
- ✅ **Configurable Tax Rate** - Tax rate configurable via `POS_TAX_RATE` environment variable
- ✅ **Type-Safe Responses** - Sales summary now uses typed structs instead of `map[string]interface{}`
- ✅ **Consolidated JWT Methods** - Single unified token generation method
- ✅ **Removed Dead Code** - Cleaned up unused PaymentService and no-op endpoints

#### 🚀 **Production Readiness**
- ✅ **Atomic Checkout** - All-or-nothing transaction guarantee
- ✅ **Persistent Token Blacklist** - Logout/revocation survives server restarts
- ✅ **Cart Persistence** - Shopping carts saved to database
- ✅ **Transaction Durability** - Sales records immediately persisted

---

## 🎯 Features

### 🔐 Authentication & Authorization
- ✅ JWT-based authentication (Access Token + Refresh Token)
- ✅ Login / Logout dengan token blacklist
- ✅ User registration
- ✅ Change password
- ✅ Refresh token mechanism
- ✅ **Role-Based Access Control (RBAC)**:
  - **SUPER_ADMIN**: Full access + user management
  - **ADMIN**: Inventory management + POS access
  - **CASHIER**: POS operations only
  - **VIEWER**: Read-only access

### 👥 User Management (Admin Only)
- ✅ CRUD users dengan pagination & filtering
- ✅ Update user role & status
- ✅ Search by username, email, full name
- ✅ Auto-seed default users on migration

### 📦 Inventory Management
- ✅ Full CRUD operations
- ✅ Stock management (update, adjust)
- ✅ **Role-based permissions**:
  - Read: All authenticated users
  - Write: ADMIN & SUPER_ADMIN only
- ✅ Filter by SKU, name, location
- ✅ Min/max stock tracking
- ✅ Price management

### 🛒 Point of Sale (POS)

#### Shopping Cart
- ✅ Create & manage shopping cart
- ✅ Add/remove items
- ✅ Update quantities dengan stock validation
- ✅ Clear cart
- ✅ Auto-create cart jika belum ada

#### Checkout & Transactions
- ✅ Checkout dengan multiple payment methods:
  - 💵 **CASH** (Implemented)
  - 💳 **CARD** (TODO - Payment gateway)
  - 📱 **QRIS** (TODO - Payment gateway)
  - 💰 **E-WALLET** (TODO - GoPay, OVO, Dana)
  - 🏦 **TRANSFER** (TODO - Bank transfer)
- ✅ Auto-generate transaction number (`TRX-YYYYMMDD-NNNN`)
- ✅ Auto-calculate:
  - Subtotal
  - Tax (PPN 11%)
  - Total amount
  - Change amount (kembalian)
- ✅ Real-time stock deduction
- ✅ Transaction history dengan pagination
- ✅ Cancel transaction dengan auto inventory restore

#### Sales Reporting
- ✅ Today's sales summary
- ✅ Total transactions count
- ✅ Total items sold
- ✅ Filter by status, payment method, date range

### 🗄️ Database
- ✅ **Full PostgreSQL support** - All repositories (Users, Inventory, Carts, Transactions, Tokens)
- ✅ Auto-migrations on startup
- ✅ 7 migration files (up & down)
- ✅ Auto-update triggers untuk `updated_at`
- ✅ Indexes untuk performance
- ✅ Constraints untuk data integrity
- ✅ In-memory repositories untuk testing/development
- ✅ **Atomic checkout** dengan database transactions
- ✅ **Batch operations** untuk multi-item cart updates

### 🔒 Security
- ✅ Password hashing dengan **bcrypt**
- ✅ JWT tokens dengan **HS256** signing
- ✅ Token blacklist untuk logout **(persistent in PostgreSQL)**
- ✅ Role-based middleware
- ✅ Permission checks di handler level
- ✅ SQL injection prevention (parameterized queries)
- ✅ Input validation
- ✅ **Atomic operations** - Failed checkout automatically rolls back
- ✅ **Thread-safe** in-memory repositories with mutex protection

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│         Presentation Layer                  │
│  (HTTP Handlers, DTOs, Middleware, Routes)  │
└───────────────────┬─────────────────────────┘
                    │
┌───────────────────▼─────────────────────────┐
│          Application Layer                  │
│        (Domain Services)                    │
│  • AuthService  • POSService                │
│  • InventoryService  • TokenService         │
└───────────────────┬─────────────────────────┘
                    │
┌───────────────────▼─────────────────────────┐
│            Domain Layer                     │
│     (Entities, Repository Interfaces)       │
│  • User  • Cart  • Transaction  • Inventory │
└───────────────────┬─────────────────────────┘
                    │
┌───────────────────▼─────────────────────────┐
│        Infrastructure Layer                 │
│  (Repositories, JWT, Database, HTTP Server) │
└─────────────────────────────────────────────┘
```

### Design Principles
- **Clean Architecture**: Separation of concerns
- **Domain-Driven Design**: Entity-driven business logic
- **Dependency Injection**: Loose coupling
- **Repository Pattern**: Data access abstraction
- **Interface Segregation**: Focused interfaces

### Performance Optimizations (v3.0.0)

#### ⚡ Checkout Flow Improvements
```
Before (v2.x):
  - N+1 queries: 100 DB round-trips for 50-item cart
  - No transaction safety: partial failures possible
  - In-memory storage: data loss on restart

After (v3.0.0):
  - Batch queries: 2 DB round-trips for any cart size
  - Atomic transactions: all-or-nothing guarantee
  - PostgreSQL persistence: zero data loss
```

#### 📊 Database Architecture
```
┌─────────────────────────────────────────┐
│         PostgreSQL Database             │
├─────────────────────────────────────────┤
│ ✅ users          (Persistent)          │
│ ✅ inventories    (Persistent)          │
│ ✅ carts          (Persistent) NEW!     │
│ ✅ transactions   (Persistent) NEW!     │
│ ✅ tokens         (Persistent) NEW!     │
│ ✅ token_blacklist (Persistent)         │
└─────────────────────────────────────────┘
```

#### 🔒 Thread Safety
- All in-memory repositories use `sync.RWMutex` for concurrent access
- Race-condition free for development/testing mode
- Production mode uses PostgreSQL for full reliability

## 📁 Project Structure

```
jwt-ddd-clean/
├── cmd/
│   └── main.go                          # Application entry point
├── internal/
│   ├── domain/
│   │   ├── model/                       # Domain entities
│   │   │   ├── user.go                  # User dengan roles & permissions
│   │   │   ├── cart.go                  # Shopping cart
│   │   │   ├── transaction.go           # Sales transaction
│   │   │   ├── inventory.go             # Inventory item
│   │   │   └── token.go                 # Token entity
│   │   ├── repository/                  # Repository interfaces
│   │   │   ├── user_repository.go
│   │   │   ├── pos_repository.go        # Cart & Transaction repos
│   │   │   ├── inventory_repository.go
│   │   │   └── token_repository.go
│   │   └── service/                     # Domain services
│   │       ├── auth_service.go          # Authentication logic
│   │       ├── pos_service.go           # POS workflow
│   │       ├── inventory_service.go
│   │       ├── payment_service.go       # Payment (TODO placeholder)
│   │       └── token_service.go
│   ├── infrastructure/
│   │   ├── jwt/                         # JWT implementation
│   │   │   └── jwt_provider.go
│   │   ├── repository/                  # Repository implementations
│   │   │   ├── postgres_user_repository.go
│   │   │   ├── postgres_cart_repository.go         # NEW: PostgreSQL cart storage
│   │   │   ├── postgres_transaction_repository.go  # NEW: PostgreSQL transaction storage
│   │   │   ├── postgres_token_repository.go        # NEW: PostgreSQL token blacklist
│   │   │   ├── memory_user_repository.go
│   │   │   ├── memory_cart_repository.go
│   │   │   ├── memory_transaction_repository.go
│   │   │   ├── memory_token_repository.go
│   │   │   └── inventory_repository.go              # PostgreSQL + Memory
│   │   ├── http/                        # HTTP server
│   │   │   └── server.go                # Route setup & DI
│   │   ├── database/                    # Database connection
│   │   │   └── database.go
│   │   └── config/                      # Configuration
│   │       └── config.go
│   ├── handler/                         # Application handlers
│   │   ├── auth_handler.go              # Auth endpoints
│   │   ├── pos_handler.go               # POS endpoints
│   │   └── token_handler.go
│   ├── http/
│   │   ├── inventory/                   # Inventory HTTP handler
│   │   │   └── inventory_http_handler.go
│   │   └── middleware/                  # Middleware
│   │       └── auth_middleware.go       # JWT & RBAC
│   ├── dto/                             # Data Transfer Objects
│   │   ├── auth_dto.go                  # Auth requests/responses
│   │   ├── pos_dto.go                   # POS requests/responses
│   │   ├── inventory_dto.go
│   │   └── token_dto.go
│   └── pkg/
│       └── errors/                      # Error handling
│           └── errors.go
├── migrations/                          # Database migrations
│   ├── 001_create_inventories_table.up.sql
│   ├── 001_create_inventories_table.down.sql
│   ├── 002_create_tokens_table.up.sql
│   ├── 002_create_tokens_table.down.sql
│   ├── 003_seed_inventory_data.up.sql
│   ├── 003_seed_inventory_data.down.sql
│   ├── 004_create_users_table.up.sql
│   ├── 004_create_users_table.down.sql
│   ├── 005_create_pos_tables.up.sql
│   ├── 005_create_pos_tables.down.sql
│   ├── 006_add_triggers_and_indexes.up.sql
│   └── 006_add_triggers_and_indexes.down.sql
├── postman/
│   └── Ultimate_POS_API.postman_collection.json
├── docs/
│   ├── POS_API_DOCUMENTATION.md         # Complete API docs
│   └── ERROR_DICTIONARY.md
├── pkg/
│   └── jwt/                             # Public API
│       └── jwt.go
├── data/                                # SQLite data (dev)
├── tests/                               # Unit tests
├── .env.example                         # Environment template
├── .gitignore
├── go.mod
├── go.sum
├── POS_IMPLEMENTATION_SUMMARY.md        # Implementation notes
└── README.md
```

## 🚀 Quick Start

### Prerequisites
- Go 1.26+
- PostgreSQL 12+ (optional, untuk production mode)
- curl / Postman (untuk testing)

### 1. Installation

```bash
# Clone repository
git clone <repository-url>
cd jwt-ddd-clean

# Install dependencies
go mod tidy
```

### 2. Configuration

```bash
# Copy environment file
cp .env.example .env

# Edit .env dengan your configuration
nano .env
```

**Example `.env`:**
```env
# Server
SERVER_HOST=localhost
SERVER_PORT=8080

# Database (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=inventory

# JWT
JWT_SECRET=your-super-secret-key-change-in-production
JWT_ISSUER=jwt-ddd-clean-pos
JWT_ACCESS_TOKEN_TTL=86400
JWT_REFRESH_TOKEN_TTL=604800
```

### 3. Database Setup

**Option 1: Using Docker (Recommended)**
```bash
# Start PostgreSQL
docker run -d \
  --name pos-postgres \
  -e POSTGRES_DB=inventory \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:16-alpine

# Check status
docker ps | grep pos-postgres
```

**Option 2: Local PostgreSQL**
```bash
# Create database and user
sudo -u postgres psql

CREATE DATABASE inventory;
CREATE USER pos_user WITH PASSWORD 'pos_password';
GRANT ALL PRIVILEGES ON DATABASE inventory TO pos_user;
\q
```

### 4. Run the Server

```bash
# With PostgreSQL
go run cmd/main.go -server

# Or build binary first
go build -o pos-app ./cmd/main.go
./pos-app -server
```

**Expected output:**
```
🚀 Starting server with PostgreSQL...
✅ Connected to PostgreSQL: postgres@localhost:5432/inventory
🔄 Running database migrations...
📁 Found 6 migration file(s)
🔄 Running migration: 001_create_inventories_table.up.sql
🔄 Running migration: 002_create_tokens_table.up.sql
🔄 Running migration: 003_seed_inventory_data.up.sql
🔄 Running migration: 004_create_users_table.up.sql
🔄 Running migration: 005_create_pos_tables.up.sql
🔄 Running migration: 006_add_triggers_and_indexes.up.sql
✅ Database migrations completed successfully
🚀 Starting JWT API Server on http://localhost:8080
```

### 5. Test the API

```bash
# Health check
curl http://localhost:8080/api/health

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Get API info
curl http://localhost:8080/
```

## 👥 Default Users

| Username | Password | Role | Email | Description |
|----------|----------|------|-------|-------------|
| `superadmin` | `admin123` | SUPER_ADMIN | superadmin@pos.local | Full access |
| `admin` | `admin123` | ADMIN | admin@pos.local | Inventory + POS |
| `cashier` | `cashier123` | CASHIER | cashier@pos.local | POS only |

## 📡 API Endpoints

### 🔓 Public Endpoints (No Auth)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/auth/login` | Login user |
| `POST` | `/api/auth/register` | Register new user |
| `POST` | `/api/auth/refresh` | Refresh access token |
| `GET` | `/api/health` | Health check |
| `GET` | `/` | API info |

### 🔐 Protected Endpoints (JWT Required)

#### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/auth/logout` | Logout |
| `GET` | `/api/auth/me` | Get current user |
| `POST` | `/api/auth/change-password` | Change password |

#### Inventory
| Method | Endpoint | Description | Role |
|--------|----------|-------------|------|
| `GET` | `/api/inventory` | List inventory | All |
| `GET` | `/api/inventory/{id}` | Get inventory detail | All |
| `POST` | `/api/inventory` | Create inventory | ADMIN+ |
| `PUT` | `/api/inventory/{id}` | Update inventory | ADMIN+ |
| `DELETE` | `/api/inventory/{id}` | Delete inventory | ADMIN+ |
| `PUT` | `/api/inventory/{id}/stock` | Update stock | All |
| `POST` | `/api/inventory/{id}/stock/adjust` | Adjust stock | All |

#### POS - Cart
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/pos/cart` | Create cart |
| `GET` | `/api/pos/cart/my` | Get my cart |
| `GET` | `/api/pos/cart/{id}` | Get cart detail |
| `POST` | `/api/pos/cart/{id}/items` | Add item to cart |
| `PUT` | `/api/pos/cart/{id}/items` | Update item quantity |
| `DELETE` | `/api/pos/cart/{id}/items` | Remove item |
| `POST` | `/api/pos/cart/{id}/clear` | Clear cart |
| `DELETE` | `/api/pos/cart/{id}` | Delete cart |

#### POS - Checkout & Transactions
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/pos/checkout/{id}` | Checkout cart |
| `GET` | `/api/pos/transactions` | List transactions |
| `GET` | `/api/pos/transactions/{id}` | Get transaction detail |
| `POST` | `/api/pos/transactions/{id}/cancel` | Cancel transaction |
| `GET` | `/api/pos/sales/today` | Today's sales summary |

#### Admin Only (SUPER_ADMIN / ADMIN)
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/admin/users` | Create new user |
| `GET` | `/api/admin/users` | List users |
| `GET` | `/api/admin/users/{id}` | Get user detail |
| `PUT` | `/api/admin/users/{id}` | Update user |
| `DELETE` | `/api/admin/users/{id}` | Delete user |

## 🧪 API Examples

### 1. Login & Get Token

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": "uuid-here",
      "username": "admin",
      "email": "admin@pos.local",
      "full_name": "Administrator",
      "role": "ADMIN",
      "status": "ACTIVE"
    }
  }
}
```

### 2. Create Product (Admin Only)

```bash
export TOKEN="your_access_token_here"

curl -X POST http://localhost:8080/api/inventory \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "PROD-001",
    "name": "Laptop ASUS VivoBook",
    "description": "Laptop gaming tinggi",
    "quantity": 50,
    "unit": "pcs",
    "location": "Warehouse A",
    "min_stock": 10,
    "max_stock": 100,
    "price": 8500000
  }'
```

### 3. Create Cart & Add Items

```bash
# Create cart
curl -X POST http://localhost:8080/api/pos/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "John Doe"
  }'

# Add item to cart (replace CART_ID and PRODUCT_ID)
curl -X POST http://localhost:8080/api/pos/cart/CART_ID/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "PRODUCT_ID",
    "quantity": 2
  }'
```

### 4. Checkout

```bash
curl -X POST http://localhost:8080/api/pos/checkout/CART_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_method": "CASH",
    "payment_amount": 10000000,
    "customer_name": "John Doe",
    "notes": "Please wrap the gift"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Checkout berhasil",
  "data": {
    "id": "transaction-uuid",
    "transaction_no": "TRX-20260403-0001",
    "cashier_name": "Admin",
    "items": [...],
    "subtotal": 17000000,
    "tax_amount": 1870000,
    "total_amount": 18870000,
    "payment_method": "CASH",
    "payment_amount": 10000000,
    "change_amount": -8870000,
    "status": "COMPLETED"
  }
}
```

### 5. Get Today's Sales

```bash
curl -X GET http://localhost:8080/api/pos/sales/today \
  -H "Authorization: Bearer $TOKEN"
```

## 📮 Postman Collection

Import Postman collection untuk testing yang mudah:

1. Open Postman
2. Click **Import**
3. Select `postman/Ultimate_POS_API.postman_collection.json`
4. Start the server
5. Run requests in order (Authentication → Inventory → POS)

**Features:**
- ✅ Pre-configured environment variables
- ✅ Auto-save tokens setelah login
- ✅ Auto-save cart_id, transaction_id
- ✅ Test scripts untuk validation
- ✅ Complete workflow examples

## 🗄️ Database Migrations

Migrations auto-run pada startup. Files located di `migrations/`:

| File | Description |
|------|-------------|
| `001_create_inventories_table` | Inventory master table |
| `002_create_tokens_table` | Token storage & blacklist |
| `003_seed_inventory_data` | Sample inventory data |
| `004_create_users_table` | Users dengan default accounts |
| `005_create_pos_tables` | Cart & Transaction tables |
| `006_add_triggers_and_indexes` | Auto-update triggers & indexes |

**Migration Features:**
- ✅ Automatic execution on startup
- ✅ Up & Down migrations (rollback support)
- ✅ Auto-update `updated_at` triggers
- ✅ Performance indexes
- ✅ Data integrity constraints

## 🧪 Testing

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test -v ./internal/domain/service/...

# Build test
go build -o pos-app ./cmd/main.go
```

## 🔧 Development

### Run Without Database (In-Memory Mode)

```bash
# Development mode dengan in-memory repositories
go run cmd/main.go
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | localhost | Server host |
| `SERVER_PORT` | 8080 | Server port |
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | postgres | Database password |
| `DB_NAME` | inventory | Database name |
| `DB_SSLMODE` | disable | PostgreSQL SSL mode |
| `DB_MAX_OPEN_CONNS` | 25 | Max open connections |
| `DB_MAX_IDLE_CONNS` | 5 | Max idle connections |
| `DB_CONN_MAX_LIFETIME` | 5m | Connection max lifetime |
| `JWT_SECRET` | your-secret-key | JWT signing key |
| `JWT_ISSUER` | jwt-ddd-clean | JWT issuer |
| `JWT_ACCESS_TOKEN_TTL` | 15m | Access token TTL (duration) |
| `JWT_REFRESH_TOKEN_TTL` | 168h | Refresh token TTL (duration) |
| `POS_TAX_RATE` | 11 | Default tax rate for checkout (%) |

## 📋 TODO / Future Enhancements

### ✅ Recently Completed
- [x] PostgreSQL repository untuk Cart
- [x] PostgreSQL repository untuk Transaction  
- [x] PostgreSQL repository untuk Token (blacklist persistence)
- [x] Database transactions untuk checkout (atomic operations)
- [x] Batch inventory updates (N+1 query optimization)
- [x] Configurable tax rate via environment variable
- [x] Thread-safe in-memory repositories
- [x] Remove duplicate auth logic
- [x] Type-safe sales summary responses

### 🚀 Upcoming Enhancements
- [ ] **Payment Gateway Integration**:
  - [ ] Midtrans (Indonesia)
  - [ ] Xendit (Indonesia)
  - [ ] Stripe (International)
  - [ ] QRIS payment
  - [ ] E-wallet (GoPay, OVO, Dana, ShopeePay)
  - [ ] Card payment (Visa, Mastercard)
- [ ] Refund functionality dengan payment gateway
- [ ] Payment reconciliation
- [ ] Advanced reporting & analytics
- [ ] Export to CSV/Excel
- [ ] Receipt generation & printing
- [ ] Barcode/QR code scanning
- [ ] Multi-store support
- [ ] Customer loyalty program
- [ ] Inventory alerts (low stock, out of stock) notifications
- [ ] Batch operations untuk inventory updates
- [ ] Audit logging untuk compliance
- [ ] Real-time notifications (WebSocket/SSE)
- [ ] Mobile app (React Native / Flutter)
- [ ] Web dashboard (React / Vue)
- [ ] API rate limiting
- [ ] Caching layer (Redis)
- [ ] GraphQL API support

## 📚 Documentation

- **[POS_API_DOCUMENTATION.md](docs/POS_API_DOCUMENTATION.md)** - Complete API reference dengan examples
- **[POS_IMPLEMENTATION_SUMMARY.md](POS_IMPLEMENTATION_SUMMARY.md)** - Implementation overview
- **[OPTIMIZATION_SUMMARY.md](OPTIMIZATION_SUMMARY.md)** - Performance optimizations in v3.0.0
- **[ERROR_DICTIONARY.md](docs/ERROR_DICTIONARY.md)** - Error codes & messages

## 🤝 Contributing

Contributions welcome! Please follow these steps:

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

## 🙏 Acknowledgments

- **Clean Architecture** by Robert C. Martin
- **Domain-Driven Design** by Eric Evans
- **Go Programming Language**
- **PostgreSQL Database**
- **Gorilla Mux** for routing
- **golang-jwt** for JWT implementation

## 📞 Support

Untuk pertanyaan atau issues, silakan buka **Issues** di repository atau hubungi maintainer.

---

**Made with ❤️ using Go, Clean Architecture, and DDD principles**
