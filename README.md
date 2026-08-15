# api-ondemand — Go Load Balancer

Um cluster de 3 servidores HTTP em Go com balanceamento de carga via **nginx**, pronto para produção.

## Arquitetura

```
                        ┌─────────────────────────────────┐
  Cliente HTTP          │        nginx :80                │
  ─────────────► :80 ──►│  (load balancer — least_conn)   │
                        └──────┬────────────┬─────────────┘
                               │            │
              ┌────────────────┼────────────┼────────────────┐
              ▼                ▼            ▼
        ┌──────────┐   ┌──────────┐   ┌──────────┐
        │ Server 1 │   │ Server 2 │   │ Server 3 │
        │  :8080   │   │  :8081   │   │  :8082   │
        └──────────┘   └──────────┘   └──────────┘
```

## Funcionalidades

### nginx (Load Balancer)
| Feature | Detalhes |
|---|---|
| Algoritmo | `least_conn` — menor número de conexões ativas |
| Health check | Passivo: `max_fails=3, fail_timeout=30s` |
| Rate Limiting | 100 req/s por IP, burst de 200 |
| Gzip | Compressão automática de JSON/JS/CSS |
| Logging | JSON estruturado com latência e upstream |
| Retry | Tenta próximo upstream em caso de erro 5xx |
| Timeouts | connect 5s, send/read 60s |

### Servidores Go
| Endpoint | Método | Descrição |
|---|---|---|
| `/` | GET | Página de boas-vindas em HTML |
| `/ping` | GET | `{"status":"ok","server":"..."}` |
| `/health` | GET | Health check — sempre retorna 200 |
| `/info` | GET | Info do servidor: nome, porta, uptime, go version, hostname |
| `/echo` | POST | Ecoa o body da requisição como JSON |
| `/metrics` | GET | Métricas: total de requests, por endpoint, latência média |

### Infraestrutura
- **Graceful Shutdown** — aguarda requests em andamento antes de parar
- **Logging estruturado** — JSON com método, path, status, latência, IP
- **Config via env vars** — `PORT` e `SERVER_NAME`
- **Docker** — imagem Alpine multi-stage (< 20MB)

## Início rápido

### Localmente (Go)
```bash
# Terminal 1
PORT=8080 SERVER_NAME="Server 1" go run ./server1/

# Terminal 2
PORT=8081 SERVER_NAME="Server 2" go run ./server2/

# Terminal 3
PORT=8082 SERVER_NAME="Server 3" go run ./server3/
```

### Com Make
```bash
make run-servers   # Inicia todos em background
make test          # Smoke test em todos os endpoints
make metrics       # Exibe métricas de cada servidor
make stop          # Para todos
```

### Com Docker Compose
```bash
make docker-up     # Build + start (inclui nginx)
make docker-logs   # Acompanhar logs
make test          # Smoke test
make docker-down   # Parar tudo
```

## Teste de carga
```bash
# Instalar hey
go install github.com/rakyll/hey@latest

# 500 requisições, 50 concorrentes
make load-test

# Ou diretamente
hey -n 1000 -c 100 http://localhost/ping
```

## Estrutura do projeto

```
api-ondemand/
├── nginx/
│   └── nginx.conf          # Config do load balancer
├── server1/
│   └── main.go             # Servidor 1 (:8080)
├── server2/
│   └── main.go             # Servidor 2 (:8081)
├── server3/
│   └── main.go             # Servidor 3 (:8082)
├── docs/
│   └── nginx.doc           # Doc de referência do nginx
├── docker-compose.yml      # Orquestração Docker
├── Dockerfile.server       # Image multi-stage dos servidores
├── Makefile                # Comandos de desenvolvimento
├── go.mod
└── README.md
```

## Configuração de ambiente

| Variável | Padrão | Descrição |
|---|---|---|
| `PORT` | `8080` | Porta de escuta do servidor |
| `SERVER_NAME` | `Server 1` | Nome exibido nas respostas |
