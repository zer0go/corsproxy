# corsproxy

Minimal open CORS proxy in Go — the upstream URL is passed directly in the request path.  
Config via env using [`caarlos0/env`](https://github.com/caarlos0/env). Logging via `log/slog`.

## Request format

```
https://corsproxy-prod.up.railway.app/{upstream-url}

# Examples:
https://corsproxy-prod.up.railway.app/http://ip-api.com/json
https://corsproxy-prod.up.railway.app/https://api.example.com/v1/users?page=2
```

## Usage

```bash
go get
PROXY_ALLOWED_HOSTS=ip-api.com,api.example.com go run .
```

## Environment variables

| Variable                  | Default                                              | Description                                                       |
|---------------------------|------------------------------------------------------|-------------------------------------------------------------------|
| `PROXY_ADDR`              | `:8080`                                              | Listen address                                                    |
| `PROXY_ALLOWED_HOSTS`     | *(empty = all allowed)*                              | Comma-separated upstream host allowlist (recommended in prod)     |
| `CORS_ALLOW_ORIGINS`      | `*`                                                  | Comma-separated origins, or `*`                                   |
| `CORS_ALLOW_METHODS`      | `GET,POST,PUT,PATCH,DELETE,OPTIONS`                  | Allowed HTTP methods                                              |
| `CORS_ALLOW_HEADERS`      | `Accept,Authorization,Content-Type,X-Requested-With` | Allowed request headers                                           |
| `CORS_MAX_AGE`            | `86400`                                              | Preflight cache duration (seconds)                                |
| `CORS_ALLOW_CREDENTIALS`  | `false`                                              | Set `true` to allow cookies/credentials                           |
| `LOG_FORMAT`              | `json`                                               | `json` or `text`                                                  |
| `LOG_LEVEL`               | `info`                                               | `debug`, `info`, `warn`, `error`                                  |

> **Warning:** leaving `PROXY_ALLOWED_HOSTS` empty makes this an open proxy.  
> Always set it in production.
