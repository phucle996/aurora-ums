# UserManagmentSystem Build Guide

Tài liệu này chỉ tập trung vào build service UMS.

Các tài liệu còn lại:

- Cài đặt: `install.md`
- Cấu hình: `config.md`
- Kiến trúc: `architech.md`

## Build Local Binary

Yêu cầu:

- Go `1.25.6`

Build nhanh:

```bash
go mod download
go build -o ./bin/server ./cmd/server
```

Build production-style (static, linux/amd64):

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/server ./cmd/server
```

Kiểm tra binary:

```bash
./bin/server
```

## Build Docker Image

Build image bằng Dockerfile production:

```bash
docker build -t aurora/ums:local -f Dockerfile .
```

Run container:

```bash
docker run --rm -it \
  --name ums \
  -p 3005:3005 \
  aurora/ums:local
```

## Build + Hot Reload (Dev)

Build image dev:

```bash
docker build -t aurora/ums-dev:local -f Dockerfile.dev .
```

Run dev mode:

```bash
docker run --rm -it \
  --name ums-dev \
  -p 3005:3005 \
  -v "$(pwd)":/app \
  aurora/ums-dev:local air -c .air.toml
```

## Build bằng Root Compose

Từ thư mục root project:

```bash
docker compose build ums-service
```

Build + chạy:

```bash
docker compose up -d --build ums-service
```
