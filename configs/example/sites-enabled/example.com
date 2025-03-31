### FRONTEND ###

upstream example_frontend {
  server 127.0.0.1:3080;
}

server {
  listen 80;
  listen 443 ssl http2;

  server_name example.com www.example.com;

  ssl_certificate     certs/example.com;
  ssl_certificate_key certs/example.com.key;
  ssl_protocols       TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
  ssl_ciphers         HIGH:!aNULL:!MD5;

  location / {
    proxy_pass http://example_frontend;
    proxy_pass_request_headers on;
  }
}

### API ###

upstream example_api {
  server 127.0.0.1:3180;
}

server {
  listen 80;
  listen 443 ssl http2;

  server_name api.example.com;

  ssl_certificate     certs/example.com;
  ssl_certificate_key certs/example.com.key;
  ssl_protocols       TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
  ssl_ciphers         HIGH:!aNULL:!MD5;

  location / {
    proxy_pass http://example_api;
    proxy_pass_request_headers on;
  }

  location /live {
    proxy_pass http://example_api;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
  }
}

### COUCHDB ###

upstream example_couchdb {
  server 127.0.0.1:5984;
}

server {
  listen 80;
  listen 443 ssl http2;

  server_name couchdb.example.com;

  ssl_certificate     certs/example.com;
  ssl_certificate_key certs/example.com.key;
  ssl_protocols       TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
  ssl_ciphers         HIGH:!aNULL:!MD5;

  return 204;

  #location / {
  #  proxy_pass http://example_couchdb;
  #  proxy_pass_request_headers on;
  #}
}
