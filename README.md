# Aurora UMS

## Install Flow

```mermaid
sequenceDiagram
    autonumber
    actor UI as Admin UI
    box rgb(244, 233, 255) Control Plane
        participant Admin as Aurora Admin
        participant ETCD as etcd
        participant PG as PostgreSQL
    end
    box rgb(232, 244, 255) Execution Plane
        participant Agent as Aurora Agent
        participant UMS as aurora-ums.service
    end

    UI->>Admin: Install UMS(module_name, agent_id, app_host)
    Admin->>ETCD: Resolve target agent registry
    Admin->>PG: Create schema + run migrations
    Admin->>ETCD: Seed runtime config + endpoint + app port
    Admin->>Admin: Generate app TLS bundle
    Admin->>Admin: Issue one-time UMS bootstrap token
    Admin->>Agent: InstallModule(bundle, env, app_host, app_port)

    Agent->>Agent: Download + verify bundle
    Agent->>Agent: Render env/systemd/nginx
    Agent->>Agent: Install binary + TLS files
    Agent->>UMS: systemctl restart aurora-ums.service

    UMS->>Admin: BootstrapModuleClient(bootstrap_token + CSR)
    Admin-->>UMS: client cert + admin CA
    UMS->>Admin: GetRuntimeBootstrap(app, psql, redis, token)
    Admin-->>UMS: typed runtime config
    UMS->>UMS: Start HTTPS server

    Agent->>UMS: Healthcheck
    alt healthy
        Agent-->>Admin: install completed
    else unhealthy
        Agent->>Agent: rollback partial install
        Agent-->>Admin: install failed
    end
```

## Runtime Paths

- `env`: `/var/lib/aurora-ums/config/ums.env`
- `app tls`: `/var/lib/aurora-ums/tls`
- `admin rpc client tls`: `/var/lib/aurora-ums/adminrpc`

## Bootstrap Phases

```mermaid
flowchart TD
    subgraph UMS[UMS Runtime]
        A[UMS process starts] --> B{AdminRPC client cert exists?}
        B -- No --> C[Generate local private key]
        C --> D[Generate CSR]
        D --> E[Read bootstrap token from env]
        B -- Yes --> L[Use existing AdminRPC client cert]
        O[Receive app + psql + redis + token config]
        P[Apply runtime config]
        Q[Start HTTPS server]
        I[Write /var/lib/aurora-ums/adminrpc/client.crt]
        J[Write /var/lib/aurora-ums/adminrpc/client.key]
        K[Write /var/lib/aurora-ums/tls/ca.crt]
    end

    subgraph Admin[Admin Control Plane]
        F[Verify Admin server TLS]
        G[BootstrapModuleClient]
        N[GetRuntimeBootstrap]
    end

    subgraph Storage[Local Runtime State]
        S1[(bootstrap token in env)]
        S2[(AdminRPC client cert)]
        S3[(AdminRPC client key)]
        S4[(Admin CA)]
    end

    E --> S1
    E --> F
    F --> G
    G --> I
    G --> J
    G --> K
    I --> S2
    J --> S3
    K --> S4
    I --> L
    J --> L
    K --> L
    L --> N
    N --> O
    O --> P
    P --> Q

    classDef ums fill:#e9f7ef,stroke:#2e8b57,stroke-width:1px,color:#113322;
    classDef admin fill:#efe7fb,stroke:#6a4fb3,stroke-width:1px,color:#25163f;
    classDef storage fill:#fff4d6,stroke:#b8860b,stroke-width:1px,color:#4a3900;

    class A,B,C,D,E,I,J,K,L,O,P,Q ums;
    class F,G,N admin;
    class S1,S2,S3,S4 storage;
```
