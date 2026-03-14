# UserManagmentSystem Architecture

## Tổng quan

UMS xử lý authentication, MFA, RBAC, device sessions.

## Runtime bootstrap

1. UMS start.
2. Nếu thiếu `AdminRPC` client cert/key:
   - tự sinh private key local
   - tạo CSR local
   - dùng one-time bootstrap token gọi `Admin`
   - nhận `client cert + admin CA`
3. UMS gọi `GetRuntimeBootstrap` qua mTLS để lấy runtime config.
4. Apply config runtime.
5. Kết nối PostgreSQL + Redis.
6. Bootstrap token secrets từ Redis cache và subscribe invalidate channel.

## Data stores

- PostgreSQL: user, profile, rbac, mfa, refresh metadata.
- Redis:
  - jwt blacklist
  - permission/device/mfa caches
  - rate limit state
  - token secret cache (do Admin rotate + publish invalidate)

## Security

- UMS chạy HTTPS với app cert riêng tại `/etc/aurora/certs/ums.crt` và `/etc/aurora/certs/ums.key`.
- Admin RPC client dùng cert riêng:
  - `/etc/aurora/certs/ums-adminrpc-client.crt`
  - `/etc/aurora/certs/ums-adminrpc-client.key`
- Private key `AdminRPC` client được sinh local trên node UMS, không đi từ Admin xuống.
- Admin RPC bootstrap dùng `bootstrap token + CSR`, sau đó chuyển sang mTLS.
