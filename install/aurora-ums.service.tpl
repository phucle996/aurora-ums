[Unit]
Description=Aurora User Management Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=aurora
Group=aurora
WorkingDirectory=/opt/aurora/ums
StateDirectory=aurora-ums
EnvironmentFile=-{{ .EnvFilePath }}
ExecStart={{ .BinaryPath }}
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
