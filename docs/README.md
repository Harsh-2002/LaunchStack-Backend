# LaunchStack Backend

LaunchStack is a platform for deploying and managing n8n instances.

## Directory Structure

- `config/`: Configuration files and structures
- `container/`: Docker container management code
- `db/`: Database models and migrations
- `docs/`: Documentation files
- `middleware/`: Middleware for authentication, CORS, etc.
- `models/`: Data models
- `routes/`: API route handlers
- `tests/`: Test scripts and tools

## Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and update with your configuration
3. Run `go build -o launchstack-backend main.go`
4. Run `./launchstack-backend`

## Testing

To run tests, use the `run_tests.sh` script:

```bash
./run_tests.sh [test_name]
```

For more information on testing, see the [tests/README.md](tests/README.md) file.

## Documentation

See the `docs/` directory for detailed documentation:

- [API Documentation](docs/API_DOCUMENTATION.md)
- [Architecture Diagram](docs/ARCHITECTURE_DIAGRAM.md)
- [Authentication Documentation](docs/AUTH_DOCUMENTATION.md)
- [Database Schema](docs/DATABASE_SCHEMA.md)
- [DNS Management](docs/DNS_MANAGEMENT.md)
- [Environment Setup](docs/ENV_SETUP.md)
- [Improvement Checklist](docs/IMPROVEMENT_CHECKLIST.md)
- [Resource Allocation](docs/RESOURCE_ALLOCATION.md)

## License

Proprietary. 