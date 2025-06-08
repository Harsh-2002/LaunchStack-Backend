# LaunchStack Tests

This directory contains various tests, tools, and scripts for testing the LaunchStack backend.

## Directory Structure

- `scripts/`: Shell scripts for testing various components of the system
- `tools/`: Go programs for testing specific functionality

## Running Tests

To run tests, use the `run_tests.sh` script in the root directory:

```bash
./run_tests.sh [test_name]
```

Available tests:
- `clerk-webhook`: Test Clerk webhook integration
- `paypal-auth`: Test PayPal authentication
- `paypal-checkout`: Test PayPal checkout flow
- `paypal-webhook`: Test PayPal webhook integration
- `docker`: Test Docker integration
- `dns`: Test DNS integration
- `cors`: Test CORS configuration
- `all`: Run all tests

## Testing Environment

Most tests require a running LaunchStack backend instance. You can start the backend using:

```bash
go build -o launchstack-backend main.go
./launchstack-backend
```

## Adding New Tests

When adding new tests:
1. Place shell scripts in the `scripts/` directory
2. Place Go programs in the `tools/` directory
3. Update the `run_tests.sh` script to include your new test 