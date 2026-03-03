# UserManagmentSystem Architecture

## 1) Mục tiêu

UMS là service quản lý định danh và phân quyền trong Aurora:

- Authentication + token lifecycle
- MFA (TOTP + recovery codes)
- RBAC (roles/permissions)
- User profile

## 2) Kiến trúc tổng thể

```text
Client (UI/Admin)
   |
   v
Gin HTTP API (handlers + middleware)
   |
   +--> Domain Services (auth/user/mfa/rbac/device/refresh/ott)
   |       |
   |       +--> Repositories (PostgreSQL)
   |       +--> Caches/blacklist (Redis)
   |
   +--> Token secret sync (etcd, bắt buộc)
```

## 3) Các lớp chính

- `cmd/server`: entrypoint khởi tạo config và vòng đời ứng dụng.
- `internal/app`: bootstrap dependency, init router, register routes, graceful shutdown.
- `internal/transport/http/handler`: nhận request/validate/response.
- `internal/transport/http/middleware`: JWT auth, permission check, rate limit, device context, CORS, access log.
- `internal/service`: business logic use-case.
- `internal/repository`: truy cập dữ liệu PostgreSQL.
- `infra/*`: adapter hạ tầng cho Postgres/Redis/etcd.

## 4) Data stores

- PostgreSQL:
  - user account, profile, RBAC, MFA, refresh token metadata.
- Redis:
  - JWT blacklist
  - device secret cache
  - permission cache
  - MFA session cache
  - rate limiting state
- etcd:
  - nguồn secret runtime cho `access_jwt`, `refresh_jwt`, `device_token`.

## 5) Request flow

1. Request vào Gin router.
2. Middleware chain chạy (context, logging, CORS, auth/rate limit/permission theo route).
3. Handler map input -> service call.
4. Service xử lý nghiệp vụ và gọi repository/cache.
5. Response chuẩn JSON trả về client.

## 6) Security controls

- JWT-based auth middleware cho protected endpoints.
- Permission middleware kiểm tra RBAC cho nhóm `/rbac`.
- Rate limit theo action key cho endpoint nhạy cảm.
- Device context để gắn device-id và hỗ trợ refresh/logout theo device.
- MFA challenge/verify cho luồng đăng nhập tăng cường.

## 7) Health & lifecycle

- `GET /health/liveness`: process sống.
- `GET /health/startup`: app đã boot xong chưa.
- `GET /health/readiness`: kiểm tra Postgres + Redis.
- Graceful shutdown:
  - mark not-ready
  - shutdown HTTP server
  - close Postgres/Redis/etcd.

## 8) Design notes

- Cấu hình đọc từ env để dễ deploy đa môi trường.
- Token secret sync từ etcd cho phép rotate secret không cần restart service.
- Router tách theo nhóm `auth` và `rbac` để kiểm soát middleware rõ ràng.
