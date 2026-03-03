# UserManagmentSystem Installation

Tài liệu này mô tả cách cài đặt và chạy UMS.

## 1) Prerequisites

- Go `1.25.6`
- Docker + Docker Compose v2
- PostgreSQL `16+`
- Redis `7+`
- etcd `v3` (nếu bật token secret sync)
- Tool migrate (`golang-migrate`)

Cài migrate (Linux):

```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.3/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
migrate -version
```

## 2) Chuẩn bị config

Trong thư mục `UserManagmentSystem`, tạo/chỉnh `.env`.

Tối thiểu cần:

```env
APP_HOST=0.0.0.0
APP_PORT=3005
DATABASE_URL=postgres://aurora:27012004@localhost:5432/aurora
REDIS_ADDR=localhost:6379
ETCD_ENDPOINTS=localhost:2379
TOKEN_SECRET_SYNC_ENABLED=false
```

Chi tiết full biến môi trường xem `config.md`.

## 3) Migrate database

```bash
cd migrations
DATABASE_URL='postgres://aurora:27012004@localhost:5432/aurora' ./scripts/migrate.sh up
cd ..
```

Kiểm tra version:

```bash
cd migrations
DATABASE_URL='postgres://aurora:27012004@localhost:5432/aurora' ./scripts/migrate.sh version
cd ..
```

## 4) Run local (không Docker)

```bash
go mod download
go run ./cmd/server
```

Hoặc hot reload:

```bash
air -c .air.toml
```

Health checks:

- `GET http://localhost:3005/health/liveness`
- `GET http://localhost:3005/health/readiness`
- `GET http://localhost:3005/health/startup`

## 5) Run bằng root docker compose

Từ root project:

```bash
docker compose up -d postgres redis etcd ums-service nginx
```

Truy cập qua nginx:

- `https://ums.aurora.local/health/liveness`
- `https://ums.aurora.local/health/readiness`

## 6) Domain local (tuỳ chọn)

Thêm vào `/etc/hosts`:

```text
127.0.0.1 aurora.local ums.aurora.local vm.aurora.local mail.aurora.local admin.aurora.local
```

## 7) Troubleshooting nhanh

- Readiness fail postgres: kiểm tra `DATABASE_URL`, DB đã chạy chưa.
- Readiness fail redis: kiểm tra `REDIS_ADDR` và auth/TLS.
- Service fail lúc boot với etcd: set `TOKEN_SECRET_SYNC_ENABLED=false` hoặc đảm bảo etcd có secret key đúng prefix.
- CORS lỗi ở browser: chỉnh `CORS_ALLOW_ORIGINS` theo đúng domain frontend.
