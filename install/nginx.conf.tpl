server {
  listen 80;
  listen [::]:80;
  server_name {{ .AppHost }};
  return 301 https://$host$request_uri;
}

server {
  listen 443 ssl;
  listen [::]:443 ssl;
  http2 on;
  server_name {{ .AppHost }};

  ssl_certificate /etc/aurora/certs/ums.crt;
  ssl_certificate_key /etc/aurora/certs/ums.key;
  ssl_protocols TLSv1.2 TLSv1.3;
  ssl_session_timeout 10m;
  ssl_prefer_server_ciphers off;

  ssl_client_certificate /etc/aurora/certs/ca.crt;
  ssl_verify_client optional;

  location / {
    proxy_pass https://127.0.0.1:{{ .AppPort }};
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto https;
    proxy_set_header Connection "";
    proxy_connect_timeout 30s;
    proxy_send_timeout 300s;
    proxy_read_timeout 300s;
    proxy_buffering off;
    proxy_request_buffering off;

    proxy_ssl_server_name on;
    proxy_ssl_name {{ .AppHost }};
    proxy_ssl_verify on;
    proxy_ssl_trusted_certificate /etc/aurora/certs/ca.crt;
    proxy_ssl_certificate /etc/aurora/certs/ums.crt;
    proxy_ssl_certificate_key /etc/aurora/certs/ums.key;
  }
}
