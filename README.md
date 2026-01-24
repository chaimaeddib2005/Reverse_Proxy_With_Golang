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

## To Do
        ✓ Implement logic to ensure a specific client (by IP or Cookie) always hits the same backend.

        ✓ Assign weights to backends based on their hardware capacity ($Weight_i$).

        ✓ Allow the proxy to serve traffic over SSL.
        


# Project description:

        # Concurrent Load-Balancing Reverse Proxy

        A production-ready reverse proxy implementation in Go featuring automatic load balancing, health monitoring, and dynamic backend management.

        ## Overview

        This reverse proxy acts as a gateway between clients and multiple backend servers, intelligently distributing incoming HTTP requests while continuously monitoring backend health. The system supports runtime configuration changes through a RESTful admin API.

        ## Features

        - **Load Balancing**: Round-robin and least-connections strategies
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

        ### Project Structure
        ```
        project/
                ├── main.go                # Entry point, orchestration
                ├── config/
                │   └── config.go          # Configuration loading
                ├── proxy/
                │   ├── handler.go         # HTTP handler logic
                │   ├──LoadBalencer.go     # Load-balancer abstract inteface
                │   ├── pool.go            # ServerPool implementation
                │   ├──backend.go          # Backend struct and methods
                |   └── sticky.go          # server pool with sticky sessions
                ├── health/
                │   └── checker.go         # Health checking logic
                ├── admin/
                │   └── api.go             # Admin API handlers
                ├── Servers/
                │   └── mock_backend.go    # backend servers for testing the proxy
                └── config.json            # Configuration file
        ```

        ## Installation

        ### Prerequisites

        - Go 1.18 or higher
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

        ## Configuration

        The `config.json` file controls all aspects of the proxy:
        ```json
        {
        "port": 8080,
        "admin_port": 8081,
        "strategy": "round-robin",
        "health_check_frequency": "30s",
        "backend_timeout": "10s",
        "health_check_method": "tcp",
        "backends": [
        "http://localhost:8082",
        "http://localhost:8083",
        "http://localhost:8084"
        ]
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
        | `backends` | array | Initial backend server URLs | Array of valid HTTP URLs |

        ## Usage

        ### Starting the Proxy
        ```bash
        go run .
        ```