# Device Assignment API Service

A secure Go-based API service for managing device authentication and user-to-device assignments using mTLS and JWT authentication.

## Features

- **Certificate-based Device Authentication**: Secure mTLS authentication for IoT devices
- **User Management**: JWT-based user authentication and authorization
- **Device Registration**: Automatic device registration upon first authentication
- **Assignment Management**: Assign/unassign devices to/from users
- **RESTful API**: Clean REST endpoints following OpenAPI standards
- **PostgreSQL Storage**: Robust data persistence with proper indexing
- **Observability**: Structured logging with JSON output
- **Security**: TLS 1.2+, proper certificate validation, secure headers

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 12+
- OpenSSL (for certificate generation)

### Installation

1. Clone and setup the project:

```bash
git clone <repository-url>
cd device-assignment-api
go mod download
```

2. Generate development certificates:

```bash
./scripts/setup-certs.sh
```

3. Set up your environment variables (copy from `env.example`):

```bash
cp env.example .env
# Edit .env with your configuration
```

4. Start PostgreSQL and create the database:

```bash
createdb device_assignment
```

5. Run the service:

```bash
export $(cat .env | grep -v ^# | xargs)
go run cmd/server/main.go
```

## API Endpoints

### Device Authentication

- `POST /api/v1/devices/authenticate` - Authenticate device with client certificate

### Device Management (JWT Required)

- `GET /api/v1/devices/{deviceId}` - Get device details
- `POST /api/v1/devices/{deviceId}/assign` - Assign device to authenticated user
- `DELETE /api/v1/devices/{deviceId}/unassign` - Unassign device from user
- `GET /api/v1/users/me/devices` - Get all devices assigned to authenticated user

### Health Check

- `GET /health` - Service health status

## Authentication

### Device Authentication (mTLS)

Devices authenticate using client certificates. The certificate's serial number and issuer are used as unique device identifiers.

Example using curl:

```bash
curl --cert ./certs/client.crt --key ./certs/client.key \
     --cacert ./certs/ca.crt \
     -X POST https://localhost:8443/api/v1/devices/authenticate
```

### User Authentication (JWT)

Users authenticate using JWT tokens in the Authorization header:

```bash
curl -H "Authorization: Bearer <jwt-token>" \
     https://localhost:8443/api/v1/users/me/devices
```

## Configuration

All configuration is done via environment variables:

| Variable         | Description                            | Default     |
| ---------------- | -------------------------------------- | ----------- |
| `SERVER_PORT`    | Server port                            | `8443`      |
| `DB_HOST`        | Database host                          | `localhost` |
| `DB_PASSWORD`    | Database password                      | _required_  |
| `TLS_CERT_FILE`  | Server certificate file                | _required_  |
| `TLS_KEY_FILE`   | Server private key file                | _required_  |
| `TLS_CA_FILE`    | CA certificate for client verification | _required_  |
| `JWT_SECRET_KEY` | JWT signing secret                     | _required_  |

See `env.example` for all available options.

## Development

### Project Structure

```
device-assignment-api/
├── cmd/server/          # Application entry point
├── internal/           # Private application code
│   ├── config/         # Configuration management
│   ├── database/       # Database layer
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models and interfaces
│   └── services/       # Business logic
├── pkg/               # Public packages
│   ├── auth/          # Authentication utilities
│   └── logger/        # Logging utilities
├── scripts/           # Development scripts
└── migrations/        # Database migrations
```

### Testing

Generate test certificates:

```bash
./scripts/setup-certs.sh
```

Test device authentication:

```bash
go run scripts/test-client.go auth
```

### Database Migrations

Database migrations run automatically on startup. Manual migration files can be placed in the `migrations/` directory.

## Security Considerations

- Uses TLS 1.2+ with strong cipher suites
- Client certificate validation with proper CA verification
- JWT tokens with configurable expiration
- SQL injection protection with parameterized queries
- Structured logging without sensitive data exposure

## Production Deployment

1. Use proper certificates from a trusted CA
2. Set strong `JWT_SECRET_KEY`
3. Configure PostgreSQL with SSL
4. Use environment-specific configuration
5. Set up proper monitoring and alerting
6. Configure log aggregation

### Docker Deployment

```bash
docker build -t device-assignment-api .
docker run -p 8443:8443 \
  -e DB_PASSWORD=secret \
  -e JWT_SECRET_KEY=your-secret \
  -v /path/to/certs:/certs \
  device-assignment-api
```

## Contributing

1. Follow the Clean Code principles outlined in the project
2. Keep functions small and focused
3. Use meaningful names for variables and functions
4. Write tests for new functionality
5. Ensure proper error handling

## License

[Your License Here]
