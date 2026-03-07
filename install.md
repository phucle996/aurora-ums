# UserManagmentSystem Installation

## 1) Prerequisites

- Go `1.26+`
- PostgreSQL `16+`
- Redis `7+`
- Admin service đang chạy gRPC/TLS

## 2) Migrate database

```bash
cd migrations
DATABASE_URL='postgres://aurora:27012004@localhost:5432/aurora' ./scripts/migrate.sh up
cd ..
```

## 3) Run local

```bash
go mod download
go run ./cmd/server
```

UMS sẽ pull runtime config từ Admin qua RPC khi startup.

## 4) Bootstrap RPC

Phải set env endpoint thật của Admin (qua domain nginx), không có fallback default.

Ví dụ:

```bash
export ADMIN_RPC_ENDPOINT=admin.aurora.local:443
```

## 5) Health check

- `GET https://ums.aurora.local/health/liveness`
- `GET https://ums.aurora.local/health/readiness`
