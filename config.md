# UserManagmentSystem Configuration

UMS không dùng etcd nữa. Runtime config được pull từ Admin service qua gRPC lúc startup.

## 1) Bootstrap config

UMS dùng endpoint bootstrap cố định:

- `admin.aurora.local:3009`

## 2) TLS paths (fixed)

- server cert: `/etc/aurora/certs/ums.crt`
- server key: `/etc/aurora/certs/ums.key`
- CA: `/etc/aurora/certs/ca.crt`
- mTLS: server yêu cầu và verify client cert (`RequireAndVerifyClientCert`)

## 3) Runtime config nhận từ Admin RPC

- App: timezone, log level, port
- PostgreSQL: url, sslmode, schema
- Redis: addr, auth, tls flags
- CORS: allow origins/methods/headers
- Token TTL: access/refresh/device/ott
- Token secret cache config: hardcoded trong service (`aurora:token-secret`, `aurora:token-secret:invalidate`, poll `10s`)

## 4) Token secret

- Admin rotate secret và cập nhật vào Redis cache.
- UMS bootstrap + subscribe invalidate channel từ Redis để cập nhật in-memory secret runtime.
