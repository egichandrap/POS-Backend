# Project Refactoring Guide

## Summary of Changes Made

### 1. Domain Layer Improvements ✅
- **Added Value Objects**: `email.go`, `password.go` in `domain/valueobject/`
- **Improved Entity Encapsulation**: All entity fields are now unexported with getter methods
- **Ubiquitous Language Methods**: Added domain-specific methods like `Activate()`, `Suspend()`, `AddStock()`, `Complete()`, etc.
- **Dual Factory Pattern**: Each entity now has `New<Entity>()` for creation and `Reconstruct<Entity>()` for database reconstitution

### 2. Application Layer Created ✅
- **DTOs Moved**: All DTOs moved from `internal/dto/` to `internal/application/dto/`
- **Use Cases Created**: Thin orchestrators in `internal/application/usecase/`
  - `auth_usecase.go`
  - `inventory_usecase.go`
  - `pos_usecase.go`
  - `token_usecase.go`

### 3. Remaining Work

#### Infrastructure Layer (`internal/infrastructure/`)
Current structure needs reorganization:
```
internal/infrastructure/
├── config/           ✅ Keep
├── database/         ✅ Keep  
├── http/             ⚠️ Move to delivery layer
├── jwt/              ✅ Keep as jwt/
└── repository/       ⚠️ Rename to persistence/
```

**Recommended Structure:**
```
internal/infrastructure/
├── config/
│   └── config.go
├── database/
│   ├── database.go
│   └── migrate.go
├── jwt/
│   └── jwt_provider.go
└── persistence/
    ├── memory_user_repository.go
    ├── postgres_user_repository.go
    ├── memory_cart_repository.go
    ├── memory_transaction_repository.go
    ├── memory_token_repository.go
    └── inventory_repository.go (postgres)
```

#### Delivery Layer (Handlers)
Current issues:
- Handlers split between `internal/handler/` and `internal/http/`
- `internal/infrastructure/http/` contains server setup (should stay)
- `internal/handler/` should become `internal/delivery/http/handlers/`

**Recommended Structure:**
```
internal/delivery/
└── http/
    ├── server.go                 (from infrastructure/http/)
    ├── middleware/
    │   └── auth_middleware.go
    └── handlers/
        ├── auth_handler.go       (update to use usecases)
        ├── inventory_handler.go  (from http/inventory/)
        ├── pos_handler.go
        └── token_handler.go
```

### 4. Required Updates

#### A. Update Domain Services
The domain services (`internal/domain/service/`) currently contain too much logic. They should be thinned to only contain:
- Cross-aggregate operations
- Stateless domain algorithms  
- Domain interfaces for external services

**Files to Review:**
- `auth_service.go` - Much logic should move to `auth_usecase.go`
- `pos_service.go` - Checkout logic is appropriate for domain service
- `inventory_service.go` - Validation should move to usecase
- `token_service.go` - Appropriate as domain service

#### B. Update Handler Imports
All handlers need to be updated to:
1. Import from `internal/application/dto` instead of `internal/dto`
2. Use usecases instead of services directly
3. Move to `internal/delivery/http/handlers/`

#### C. Update Server Wiring
`internal/infrastructure/http/server.go` needs to:
1. Import usecases from `internal/application/usecase`
2. Import handlers from `internal/delivery/http/handlers`
3. Wire dependencies properly: Infrastructure → Domain Interfaces → Application

#### D. Fix Breaking Changes
Since entity fields are now unexported, several files need updates:
- Repository implementations accessing entity fields
- Services using entity fields directly
- DTOs using old accessor patterns

### 5. Build Order

After all changes:
1. Fix domain layer compilation errors
2. Fix application layer imports
3. Fix infrastructure layer
4. Fix delivery/handler layer
5. Update main.go wiring
6. Run `go build ./...` to verify

### 6. Testing Strategy

1. Ensure all existing tests still pass
2. Update tests to use new accessor methods
3. Add tests for new usecase layer
4. Verify domain entity methods work correctly

## Next Steps

The most critical remaining tasks are:
1. **Rename infrastructure/repository → infrastructure/persistence**
2. **Move handlers to internal/delivery/http/**
3. **Update server.go to wire usecases**
4. **Fix all compilation errors from entity encapsulation**
5. **Test and verify build**

Due to the extensive nature of these changes, it's recommended to:
- Commit after each major section completes
- Run `go build` frequently to catch errors early
- Update tests alongside implementation changes
