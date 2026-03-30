# corsproxy

Minimal open CORS proxy in Go — the upstream URL is passed directly in the request path.  
Config via env using [`caarlos0/env`](https://github.com/caarlos0/env). Logging via `log/slog`.

## Request format

```
https://proxy.com/{upstream-url}

# Examples:
https://proxy.com/https://ipinfo.io/json
https://proxy.com/https://api.example.com/v1/users?page=2
```

## Usage

```bash
go get
PROXY_ALLOWED_HOSTS=ipinfo.io,api.example.com go run .
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

## Docker

```bash
docker build -t corsproxy .
docker run \
  -e PROXY_ALLOWED_HOSTS=ipinfo.io \
  -e CORS_ALLOW_ORIGINS=https://app.example.com \
  -p 8080:8080 corsproxy
```

> **Warning:** leaving `PROXY_ALLOWED_HOSTS` empty makes this an open proxy.  
> Always set it in production.
