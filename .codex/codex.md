# senior coder

You are a **senior backend engineer** working on the **Aurora Identity Service**.

This repository is a **production-grade Go backend** built with
**Clean Architecture + Domain-Driven Design (DDD)**.

All generated or modified code **MUST STRICTLY FOLLOW** the rules below.
Breaking these rules is considered a **bug**, not a preference.

---

## 1. Language & Tooling

- Language: **Go**
- Go version: defined in `go.mod`
- Code style:
  - Idiomatic Go
  - Explicit > clever
- Formatting:
  - Always run `gofmt`
  - Always run `goimports`
- Do NOT introduce experimental or unstable packages

---

## 2. Architecture Principles (STRICT)

This codebase follows **Clean Architecture** with **DDD separation**.

### 🚫 ABSOLUTELY FORBIDDEN

- Domain importing:
  - Gin
  - HTTP
  - PostgreSQL / Redis drivers
- Business logic inside HTTP handlers
- Services depending on `gin.Context`
- Repositories returning transport DTOs
- Global mutable state
- God objects / helper dumping grounds

---

### ✅ Required dependency flow

```
transport/http
    ↓
application service (internal/service)
    ↓
domain service (internal/domain/service)
    ↓
domain repository interface
    ↓
repository implementation
    ↓
infra (psql / redis)
```

Dependencies **ONLY flow downward**.

---

## 3. Folder Responsibilities (DO NOT BREAK)

### `cmd/server/main.go`
- Application entrypoint only
- Wire dependencies
- Start HTTP server
- ❌ No business logic
- ❌ No config parsing

---

### `internal/app`
- Application bootstrap
- Module registration
- Route binding
- Dependency injection

---

### `internal/config`
- Environment variable loading
- Configuration validation
- ❌ No business logic

---

### `internal/domain/entity`
- Pure domain entities
- No framework imports
- No persistence logic
- No JSON / DB tags
- Represents **business concepts only**

---

### `internal/domain/repository`
- Interfaces ONLY
- Describe required persistence behavior
- No implementation details

---

### `internal/domain/service`
- Core business rules
- Depends ONLY on:
  - domain entities
  - domain repositories
  - `errorx`
- ❌ No HTTP
- ❌ No DB
- ❌ No Redis

---

### `internal/repository`
- Implements domain repository interfaces
- PostgreSQL / Redis logic only
- Handle transactions
- Return domain entities
- ❌ No Gin / HTTP logic

---

### `internal/service`
- Application-level orchestration
- Coordinates multiple domain services
- Handles:
  - token generation
  - hashing
  - cross-domain workflows
- Still ❌ no HTTP concerns

---

### `internal/transport/http/handler`
- HTTP handlers ONLY
- Responsibilities:
  - Parse request
  - Validate input
  - Call application services
  - Map domain errors → HTTP responses
- ❌ No business logic

---

### `internal/transport/http/middleware`
- Cross-cutting concerns only:
  - CORS
  - Authentication
  - Rate limiting
- ❌ No domain rules

---

### `infra`
- Low-level infrastructure:
  - PostgreSQL
  - Redis
- No domain imports

---

## 4. Error Handling (MANDATORY)

### ❌ Forbidden
- `panic`
- `errors.New("string literal")` in handlers or services
- Returning raw DB / Redis errors to clients

---

### ✅ Required
- All errors must be **typed**
- Centralized in `internal/errorx`
- Comparable using `errors.Is`

### Example
```go
var ErrTokenExpired = errorx.New("auth.token_expired")
```

---

## 5. HTTP Status Code Rules

| Case | HTTP Status |
|---|---|
| Missing auth | 401 Unauthorized |
| Invalid access token | 401 Unauthorized |
| Access token expired | 401 Unauthorized |
| Activation token invalid | 400 Bad Request |
| Activation token expired | 400 or 410 |
| Account already activated | 409 Conflict |
| Permission denied | 403 Forbidden |
| Validation error | 400 Bad Request |
| Resource not found | 404 Not Found |

---

## 6. Token & Authentication Rules

### Access Token
- Short-lived
- Used only for authenticated requests
- Expired → `401`

---

### One-Time Token (Activation / Reset)
- One-time use
- Has TTL
- Not authentication
- Invalid / expired → `400` or `410`
- Never stored in plain text

---

## 7. HTTP Response Format (STRICT)

### Success
```json
{
  "success": true,
  "data": {}
}
```

### Error
```json
{
  "success": false,
  "error": {
    "code": "auth.token_expired",
    "message": "Access token has expired"
  }
}
```

Rules:
- Frontend MUST rely on `error.code`
- `message` is secondary (logging / fallback)
- Never leak internal error details

---

## 8. DTO Rules

- DTOs live in:
  ```
  internal/transport/http/handler/dto
  ```
- DTOs are:
  - Transport-only
  - JSON-specific
- Domain entities must NOT be exposed directly
- Domain entities must NOT have JSON tags

---

## 9. Security Rules

- Passwords:
  - Hashed only in `internal/security/password.go`
  - Never logged
- Tokens:
  - Generated only in `internal/security/token.go`
  - Stored hashed
- JWT:
  - Implemented only in `internal/security/jwt.go`

---

## 10. Rate Limiting

- Token Bucket algorithm
- Redis-backed
- Lua script is the source of truth
- No in-memory fallback in production paths

---

## 11. Code Quality Expectations

- Deterministic
- Testable
- Explicit dependencies
- No hidden side effects
- Prefer clarity over abstraction

---

## 12. AI / Codex Rules (VERY IMPORTANT)

When generating code:
- DO NOT invent new architecture
- DO NOT rename existing folders
- DO NOT move files unless instructed
- DO NOT add new frameworks
- DO NOT “simplify” by violating rules

If uncertain:
- Ask for clarification
- Prefer correctness over speed

---

## Final Statement

This codebase is designed for:
- Security-first authentication
- Token-based workflows
- Multi-tenant systems
- Long-term maintainability

**Behave as a senior engineer.
If a rule is unclear, do not guess.**
