# ngonx

**ngonx** is a HTTP server and reverse proxy written in Go.
It aims to be a drop-in replacement for NGINX with compatibility for existing NGINX configurations.

## Features

- **Full NGINX Configuration Compatibility**: Use your existing NGINX config files with minimal modification
- **Enhanced Performance**: Built on Go's efficient concurrency model for better resource utilization
- **Single Binary**: No dependencies, simple to deploy and upgrade
- **Dynamic Modules**: [planned] Extensible architecture using Go plugins
- **Live Configuration Reloading**: [planned] Zero downtime configuration changes
- **API-First Design**: [planned] REST API for configuration and monitoring
- **Observability**: [planned] Prometheus metrics, structured logging, and tracing

### Basic Usage

1. Use or create your existing NGINX configuration:

```nginx
# /etc/nginx/nginx.conf
http {
    server {
        listen 80;
        server_name example.com;

        location / {
            root /var/www/html;
            index index.html;
        }
    }
}
```

2. Start ngonx with your configuration:

```bash
ngonx -c /etc/nginx/nginx.conf
```

3. Test your server:

```bash
curl http://localhost
```

## Migration from NGINX

ngonx is designed to be a drop-in replacement for NGINX. In most cases, you can simply:

1. Install ngonx
2. Point it to your existing NGINX configuration file
3. Start ngonx instead of NGINX

### Key Configuration Options

```nginx
# Global settings
error_log /var/log/ngonx/error.log;
pid /var/run/ngonx.pid;

events {
    worker_connections 4096;
    use epoll;
}

http {
    include mime.types;
    default_type application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                     '$status $body_bytes_sent "$http_referer" '
                     '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/ngonx/access.log main;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;

    keepalive_timeout 65;

    gzip on;

    server {
        listen 80;
        server_name example.com;

        location / {
            root /var/www/html;
            index index.html;
        }
    }
}
```

## REST API [PLANNED]

ngonx includes a REST API for dynamic configuration and monitoring. Enable it with:

```nginx
http {
    # Other directives...

    server {
        listen 8080;

        location /api/ {
            api enable;
            api_auth_key "your-secret-key";
        }
    }
}
```

Example API usage:

```bash
# Get current status
curl -H "Authorization: Bearer your-secret-key" http://localhost:8080/api/status

# Update a server block
curl -X PUT -H "Authorization: Bearer your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"listen": 8081, "server_name": "api.example.com"}' \
  http://localhost:8080/api/http/servers/0
```

## Plugins and Extensions [PLANNED]

ngonx supports plugin system:

```go
package main

import "github.com/CodePortingInitiative/ngonx/plugin"

type SamplePlugin struct{}

func (p *SamplePlugin) OnRequest(ctx *plugin.Context) error {
    // Add a header to the request
    ctx.Request.Header.Set("X-Processed-By", "ngonx-sample-plugin")
    return nil
}

// Export the plugin
var Plugin = &SamplePlugin{}
```

Build and use the plugin:

```bash
go build -buildmode=plugin -o sample_plugin.so sample_plugin.go

# Enable in configuration
plugins {
    load "path/to/sample_plugin.so";
}
```

## Observability

### Metrics [PLANNED]

ngonx exports Prometheus metrics by default on port 9091:

```nginx
http {
    metrics {
        listen 9091;
        prefix "ngonx_";
        collect ["connections", "http", "cache"];
    }
}
```

### Logging [PLANNED]

ngonx supports structured JSON logging:

```nginx
error_log /var/log/ngonx/error.log json info;
access_log /var/log/ngonx/access.log json;
```

## Contributing

I welcome contributions! Please contact me [contact@anarjafarov.me](mailto:contact@anarjafarov.me) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/CodePortingInitiative/ngonx.git
cd ngonx

# Install development dependencies
make setup

# Run tests
make test

# Build for development
make build

# Run with a test configuration
./bin/ngonx -c ./configs/test.conf
```

---

© 2025 ngonx Project Contributors
