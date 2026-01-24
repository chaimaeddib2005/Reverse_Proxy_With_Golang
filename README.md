# Tasks:
## Done
✓ Define data structures (Backend, ServerPool, ProxyConfig)

✓ Implement LoadBalancer interface and ServerPool methods

✓ Create configuration loading logic

✓ Build basic proxy handler with request forwarding

✓ Add connection counting with atomic operations

✓ Implement health checker goroutine

✓ Create Admin API endpoints

✓ Add graceful shutdown handling

✓ Implement error handling and logging

✓ Test with multiple backend servers

✓ Implement logic to ensure a specific client (by IP or Cookie) always hits the same backend.

✓ Assign weights to backends based on their hardware capacity ($Weight_i$).

✓ Allow the proxy to serve traffic over SSL.

## To Do

:) :) :)

# Project description:

# Concurrent Load-Balancing Reverse Proxy

A production-ready reverse proxy implementation in Go featuring automatic load balancing, health monitoring, dynamic backend management, sticky sessions, weighted distribution, and SSL/TLS support.

## Overview

This reverse proxy acts as a gateway between clients and multiple backend servers, intelligently distributing incoming HTTP requests while continuously monitoring backend health. The system supports runtime configuration changes through a RESTful admin API.

## Features

- **Load Balancing**: Round-robin, weighted round-robin, and least-connections strategies
- **Sticky Sessions**: Client IP-based session persistence with configurable TTL
- **Weighted Backends**: Distribute traffic based on backend server capacity
- **SSL/TLS Support**: Secure HTTPS connections with certificate configuration
- **Health Monitoring**: Automatic backend health checks with configurable intervals
- **Dynamic Management**: Add or remove backends without restart via Admin API
- **Graceful Shutdown**: Clean termination of in-flight requests
- **Connection Tracking**: Real-time monitoring of active connections per backend
- **Thread-Safe Operations**: Concurrent request handling with mutex protection
- **Configurable Timeouts**: Customizable backend and health check timeouts
- **Multiple Health Check Methods**: TCP or HTTP-based health verification

## Architecture

### Components

1. **Proxy Core**: Intercepts and forwards HTTP requests using configurable load balancing strategies
2. **Health Checker**: Background service that periodically verifies backend availability
3. **Admin API**: REST endpoints for runtime configuration and status monitoring
4. **Configuration Manager**: JSON-based configuration with validation
5. **Sticky Session Manager**: Maintains client-to-backend mappings with automatic cleanup

### Project Structure
```
project/
        ├── main.go                # Entry point, orchestration
        ├── config/
        │   └── config.go          # Configuration loading
        ├── proxy/
        │   ├── handler.go         # HTTP handler logic
        │   ├── LoadBalancer.go    # Load-balancer abstract interface
        │   ├── pool.go            # ServerPool implementation
        │   ├── backend.go         # Backend struct and methods
        │   └── sticky.go          # Server pool with sticky sessions
        ├── health/
        │   └── checker.go         # Health checking logic
        ├── admin/
        │   └── api.go             # Admin API handlers
        ├── Servers/
        │   └── mock_backend.go    # Backend servers for testing the proxy
        ├── certs/
        │   ├── server.crt         # SSL certificate (optional)
        │   └── server.key         # SSL private key (optional)
        └── config.json            # Configuration file
```

## Installation

### Prerequisites

- Go 1.18 or higher
- OpenSSL (for generating SSL certificates, optional)
- Basic understanding of HTTP and networking concepts

### Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd reverseproxy
```

2. Initialize Go module:
```bash
go mod init reverseproxy.com
go mod tidy
```

3. Create configuration file:
```bash
cp config.example.json config.json
```

4. (Optional) Generate SSL certificates for HTTPS:
```bash
mkdir certs
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost"
```

## Configuration

The `config.json` file controls all aspects of the proxy:

```json
{
    "port": 8080,
    "admin_port": 8081,
    "strategy": "round-robin",
    "health_check_frequency": "30s",
    "health_check_method": "http",
    "backend_timeout": "10s",
    "backends": [
        {
            "url": "http://localhost:8082",
            "weight": 5
        },
        {
            "url": "http://localhost:8083",
            "weight": 3
        },
        {
            "url": "http://localhost:8084",
            "weight": 2
        }
    ],
    "enable_sticky_sessions": false,
    "sticky_session_ttl": "1h",
    "ssl": {
        "enabled": false,
        "cert_file": "./certs/server.crt",
        "key_file": "./certs/server.key"
    }
}
```

### Configuration Parameters

| Parameter | Type | Description | Valid Values |
|-----------|------|-------------|--------------|
| `port` | integer | Main proxy server port | 1-65535 |
| `admin_port` | integer | Admin API port | 1-65535 (must differ from port) |
| `strategy` | string | Load balancing algorithm | "round-robin", "least-conn" |
| `health_check_frequency` | string | Health check interval | Duration string (e.g., "30s", "1m") |
| `backend_timeout` | string | Backend request timeout | Duration string (e.g., "10s") |
| `health_check_method` | string | Health verification method | "tcp", "http" |
| `backends` | array | Backend server configurations | Array of objects with `url` and `weight` |
| `backends[].url` | string | Backend server URL | Valid HTTP/HTTPS URL |
| `backends[].weight` | integer | Traffic weight (higher = more traffic) | Positive integer (default: 1) |
| `enable_sticky_sessions` | boolean | Enable client IP-based session persistence | true, false |
| `sticky_session_ttl` | string | Session persistence duration | Duration string (e.g., "30m", "1h") |
| `ssl.enabled` | boolean | Enable HTTPS on proxy | true, false |
| `ssl.cert_file` | string | Path to SSL certificate file | Valid file path |
| `ssl.key_file` | string | Path to SSL private key file | Valid file path |

### Load Balancing Strategies

#### Round-Robin (with Weights)
Distributes requests sequentially across backends, respecting weight values:
- Backend with weight 5 receives ~50% of traffic
- Backend with weight 3 receives ~30% of traffic
- Backend with weight 2 receives ~20% of traffic

#### Least-Connections
Routes requests to the backend with fewest active connections, ideal for long-running requests.

#### Sticky Sessions
When enabled, ensures the same client IP always connects to the same backend (until session expires or backend fails).

## Usage

### Starting Backend Servers

Start multiple backend servers on different ports for testing:

```bash
# Terminal 1 - Start backend on port 8082
go run ./Servers 8082

# Terminal 2 - Start backend on port 8083
go run ./Servers 8083

# Terminal 3 - Start backend on port 8084
go run ./Servers 8084
```

### Starting the Proxy

In a new terminal:

```bash
go run .
```

You should see output like:
```
Added backend: http://localhost:8082 (weight: 5)
Added backend: http://localhost:8083 (weight: 3)
Added backend: http://localhost:8084 (weight: 2)
The number of backend servers is: 3
Using weighted round-robin load balancing
Proxy server starting on http://:8080
2026/01/24 15:26:16 Proxy server listening on http://:8080
2026/01/24 15:26:16 Admin API listening on :8090
2026/01/24 15:26:16 Health Checker started (interval: 30s, timeout: 10s)
```

### Testing the Proxy

#### HTTP Mode
```bash
curl http://localhost:8080/
```

#### HTTPS Mode (with SSL enabled)
```bash
curl -k https://localhost:8080/
```

The `-k` flag bypasses certificate validation for self-signed certificates.

### Admin API Endpoints

The Admin API runs on the configured `admin_port` and provides management capabilities:

#### Get All Backends
```bash
curl http://localhost:8090/backends
```

#### Get Specific Backend
```bash
curl http://localhost:8090/backends/0
```

#### Add New Backend
```bash
curl -X POST http://localhost:8090/backends \
  -H "Content-Type: application/json" \
  -d '{"url":"http://localhost:8085","weight":1}'
```

#### Remove Backend
```bash
curl -X DELETE http://localhost:8090/backends/0
```

#### Update Backend Status
```bash
curl -X PUT http://localhost:8090/backends/0/status \
  -H "Content-Type: application/json" \
  -d '{"alive":false}'
```

## Advanced Features

### Sticky Sessions

Enable sticky sessions to ensure clients consistently reach the same backend:

```json
{
    "enable_sticky_sessions": true,
    "sticky_session_ttl": "1h"
}
```

Sessions are tracked by client IP address and automatically expire after the configured TTL.

### Weighted Load Balancing

Distribute traffic proportionally based on backend capacity:

```json
{
    "backends": [
        {"url": "http://powerful-server:8082", "weight": 10},
        {"url": "http://medium-server:8083", "weight": 5},
        {"url": "http://small-server:8084", "weight": 1}
    ]
}
```

### SSL/TLS Configuration

Enable HTTPS for secure client connections:

```json
{
    "ssl": {
        "enabled": true,
        "cert_file": "./certs/server.crt",
        "key_file": "./certs/server.key"
    }
}
```

For production, use certificates from a trusted Certificate Authority like Let's Encrypt.

## Monitoring and Debugging

### Health Check Logs
Monitor backend health status in real-time:
```
2026/01/24 15:26:46 Health check: http://localhost:8082 is alive
2026/01/24 15:26:46 Health check: http://localhost:8083 is alive
2026/01/24 15:26:46 Health check: http://localhost:8084 is alive
```

### Connection Tracking
View active connections per backend via the Admin API:
```bash
curl http://localhost:8090/backends | jq
```

### Request Logging
The proxy logs all forwarded requests and errors for debugging.

## Graceful Shutdown

The proxy handles shutdown signals (SIGINT, SIGTERM) gracefully:

```bash
# Press Ctrl+C or send SIGTERM
^C
2026/01/24 15:30:00 Shutdown signal received
2026/01/24 15:30:00 Proxy server stopped gracefully
2026/01/24 15:30:00 Admin server stopped gracefully
2026/01/24 15:30:00 All servers stopped
```

In-flight requests are allowed to complete within a 30-second timeout window.

## Performance Considerations

- **Concurrent Requests**: Handles multiple simultaneous connections using goroutines
- **Lock Contention**: Minimized through read-write mutexes and atomic operations
- **Health Checks**: Run asynchronously without blocking request handling
- **Connection Pooling**: Go's HTTP client automatically manages connection pools to backends

## Troubleshooting

### Common Issues

1. **404 Not Found**: Ensure backend servers are running and URLs are correct in config.json
2. **503 Service Unavailable**: All backends are down or unreachable
3. **Port Already in Use**: Check that `port` and `admin_port` are not used by other services
4. **SSL Certificate Errors**: Verify cert and key files exist and are valid

### Debug Mode

Add logging to see detailed request routing:
```go
log.Printf("Received request: %s %s", r.Method, r.URL.Path)
log.Printf("Selected backend: %s", backend.URL.String())
```

## License

[Your License Here]

## Contributing

Contributions are welcome! Please submit pull requests or open issues for bugs and feature requests.