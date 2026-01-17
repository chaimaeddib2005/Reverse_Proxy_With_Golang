## Tasks:
# Done
 ✓ Define data structures (Backend, ServerPool, ProxyConfig)

 ✓ Implement LoadBalancer interface and ServerPool methods
 
 ✓ Create configuration loading logic
# To Do
 
 ✓ Build basic proxy handler with request forwarding

 ✓ Add connection counting with atomic operations
 
 ✓ Implement health checker goroutine
 
 ✓ Create Admin API endpoints
 
 ✓ Add graceful shutdown handling
 
 ✓ Implement error handling and logging
 
 ✓ Test with multiple backend servers

## Temporary project structure

        project/
        ├── main.go                # Entry point, orchestration
        ├── config/
        │   └── config.go          # Configuration loading
        ├── proxy/
        │   ├── handler.go         # HTTP handler logic
        │   ├──LoadBalencer.go     # Load-balancer abstract inteface
        │   ├── pool.go            # ServerPool implementation
        │   └── backend.go         # Backend struct and methods
        ├── health/
        │   └── checker.go         # Health checking logic
        ├── admin/
        │   └── api.go             # Admin API handlers
        └── config.json            # Configuration file