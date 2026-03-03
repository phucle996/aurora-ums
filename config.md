# UserManagmentSystem Configuration

Service đọc config từ `.env` (nếu có), sau đó đọc từ environment variables.

## 1) App

- `APP_NAME` (default: `Aurora Cloud`)
- `APP_HOST` (default: rỗng)
- `APP_PORT` (default: `3005`)
- `APP_LOG_LEVEL` (default: rỗng)
- `APP_TIMEZONE` (default: `UTC`)

## 2) Database (PostgreSQL)

- `DATABASE_URL` (default: `postgres://aurora:27012004@localhost:5432/aurora`)
- `DB_SCHEMA` (default: `ums`)
- `DB_SSLMODE` (default: `disable`)
- `DB_SSLROOTCERT` (optional)
- `DB_SSLKEY` (optional)
- `DB_SSLCERT` (optional)

## 3) Redis

- `REDIS_ADDR` (default: `localhost:6379`)
- `REDIS_USERNAME` (optional)
- `REDIS_PASSWORD` (optional)
- `REDIS_DB` (default: `0`)
- `REDIS_TLS` (default: `false`)
- `REDIS_TLS_CA` (optional)
- `REDIS_TLS_KEY` (optional)
- `REDIS_TLS_CERT` (optional)
- `REDIS_TLS_INSECURE` (default: `false`)

## 4) etcd

- `ETCD_ENDPOINTS` (default: `localhost:2379`, hỗ trợ danh sách phân tách bằng dấu phẩy)
- `ETCD_AUTO_SYNC_INTERVAL` (default: `5m`)
- `ETCD_DIAL_TIMEOUT` (default: `5s`)
- `ETCD_DIAL_KEEPALIVE_TIME` (default: `30s`)
- `ETCD_DIAL_KEEPALIVE_TIMEOUT` (default: `10s`)
- `ETCD_USERNAME` / `ETCD_PASSWORD` (optional)
- `ETCD_TLS` (default: `false`)
- `ETCD_TLS_CA` / `ETCD_TLS_KEY` / `ETCD_TLS_CERT` (optional)
- `ETCD_TLS_SERVER_NAME` (optional)
- `ETCD_TLS_INSECURE` (default: `false`)
- `ETCD_PERMIT_WITHOUT_STREAM` (default: `false`)
- `ETCD_REJECT_OLD_CLUSTER` (default: `false`)
- `ETCD_MAX_CALL_SEND_MSG_SIZE` (default: `2097152`)
- `ETCD_MAX_CALL_RECV_MSG_SIZE` (default: `2097152`)
- `ETCD_SASL_ENABLE` (default: `false`)
- `ETCD_SASL_MECHANISM` (default: `PLAIN`)
- `ETCD_SASL_USERNAME` / `ETCD_SASL_PASSWORD` (optional)

## 5) Tokens

- `ACCESS_TOKEN_TTL` (default: `15m`)
- `REFRESH_TOKEN_TTL` (default: `168h`)
- `OTT_TTL` (default: `15m`)

Lưu ý:

- Secret không đọc từ env.
- Secret được bootstrap + watch từ etcd theo prefix `TOKEN_SECRET_SYNC_PREFIX`:
  - `access_jwt`
  - `refresh_jwt`
  - `device_token`
- `OTT` secret được derive trong runtime từ `access_jwt` đã sync, không cần `OTT_SECRET`.

## 6) Token Secret Sync (etcd -> runtime)

- `TOKEN_SECRET_SYNC_ENABLED` (default: `true`)
- `TOKEN_SECRET_SYNC_PREFIX` (default: `/admin/token-secret`)
- `TOKEN_SECRET_SYNC_BOOTSTRAP_TIMEOUT` (default: `5s`)

Lưu ý:

- UMS yêu cầu `TOKEN_SECRET_SYNC_ENABLED=true`.
- Nếu bootstrap secret từ etcd lỗi thì service sẽ fail startup.

## 7) CORS

- `CORS_ALLOW_ORIGINS`
- `CORS_ALLOW_METHODS` (default: `GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS`)
- `CORS_ALLOW_HEADERS` (default: `Origin,Content-Type,Accept,Authorization`)
- `CORS_EXPOSE_HEADERS` (default: rỗng)
- `CORS_ALLOW_CREDENTIALS` (default: `true`)
- `CORS_MAX_AGE` (default: `12h`)

## 8) Sample .env (dev)

```env
APP_NAME=Aurora Cloud
APP_HOST=0.0.0.0
APP_PORT=3005
APP_LOG_LEVEL=info
APP_TIMEZONE=UTC

DATABASE_URL=postgres://aurora:27012004@localhost:5432/aurora
DB_SCHEMA=ums
DB_SSLMODE=disable

REDIS_ADDR=localhost:6379
REDIS_DB=0
REDIS_TLS=false

ETCD_ENDPOINTS=localhost:2379
TOKEN_SECRET_SYNC_ENABLED=true

ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=240h
OTT_TTL=15m

CORS_ALLOW_ORIGINS=http://localhost:5173,http://127.0.0.1:5173
CORS_ALLOW_CREDENTIALS=true
```
