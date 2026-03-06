# UserManagmentSystem Architecture

## Tổng quan

UMS xử lý authentication, MFA, RBAC, device sessions.

## Runtime bootstrap

1. UMS start.
2. Pull runtime config từ Admin service qua gRPC.
3. Apply config runtime.
4. Kết nối PostgreSQL + Redis.
5. Bootstrap token secrets từ Redis cache và subscribe invalidate channel.

## Data stores

- PostgreSQL: user, profile, rbac, mfa, refresh metadata.
- Redis:
  - jwt blacklist
  - permission/device/mfa caches
  - rate limit state
  - token secret cache (do Admin rotate + publish invalidate)

## Security

- UMS chạy HTTPS với cert fixed path `/etc/aurora/certs`.
- Admin RPC dùng TLS.

