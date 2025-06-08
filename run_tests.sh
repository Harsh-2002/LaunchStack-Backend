#!/bin/bash

# Script to easily run tests from their organized locations

function show_help {
    echo "LaunchStack Test Runner"
    echo "Usage: ./run_tests.sh [test_name]"
    echo ""
    echo "Available tests:"
    echo "  clerk-webhook       - Test Clerk webhook integration"
    echo "  paypal-auth         - Test PayPal authentication"
    echo "  paypal-checkout     - Test PayPal checkout flow"
    echo "  paypal-webhook      - Test PayPal webhook integration"
    echo "  docker              - Test Docker integration"
    echo "  dns                 - Test DNS integration"
    echo "  cors                - Test CORS configuration"
    echo "  all                 - Run all tests"
    echo ""
    echo "Examples:"
    echo "  ./run_tests.sh paypal-auth"
    echo "  ./run_tests.sh clerk-webhook"
}

# Load environment variables
export $(grep -v '^#' .env | xargs)

# If no arguments are provided, show help
if [ $# -eq 0 ]; then
    show_help
    exit 0
fi

case "$1" in
    clerk-webhook)
        echo "Running Clerk webhook test..."
        (cd tests/scripts && ./test_clerk_webhook.sh)
        ;;
    paypal-auth)
        echo "Running PayPal authentication test..."
        (cd tests/tools && go run test_paypal_auth.go)
        ;;
    paypal-checkout)
        echo "Running PayPal checkout test..."
        (cd tests/scripts && ./test_paypal_checkout.sh)
        ;;
    paypal-webhook)
        echo "Running PayPal webhook test..."
        (cd tests/scripts && ./test_paypal_webhook.sh)
        ;;
    docker)
        echo "Running Docker integration test..."
        (cd tests/tools && go run test_docker.go)
        ;;
    dns)
        echo "Running DNS integration test..."
        (cd tests/tools && go run test_dns_delete.go)
        ;;
    cors)
        echo "Running CORS configuration test..."
        (cd tests/tools && go run test_cors.go)
        ;;
    all)
        echo "Running all tests..."
        (cd tests/tools && go run test_cors.go)
        (cd tests/tools && go run test_docker.go)
        (cd tests/tools && go run test_paypal_auth.go)
        (cd tests/scripts && ./test_clerk_webhook.sh)
        (cd tests/scripts && ./test_paypal_webhook.sh)
        ;;
    *)
        echo "Unknown test: $1"
        show_help
        exit 1
        ;;
esac 