# LaunchStack Database Schema

This document details the database schema for the LaunchStack backend, explaining what information is stored in each table and how the data is used within the application.

## Database Tables

LaunchStack uses PostgreSQL as its database with the following tables:

### 1. Users Table

Stores information about registered users and their subscription details.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_user_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    plan VARCHAR(20) DEFAULT 'starter', -- 'starter', 'pro'
    paypal_customer_id VARCHAR(255),
    subscription_id VARCHAR(255),
    subscription_status VARCHAR(50), -- 'trial', 'active', 'canceled', 'expired'
    trial_start_date TIMESTAMP,
    trial_end_date TIMESTAMP,
    current_period_end TIMESTAMP,
    billing_cycle VARCHAR(10), -- 'monthly', 'yearly'
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

**Key Fields:**
- `id`: Internal UUID for the user
- `clerk_user_id`: External ID from Clerk authentication service
- `email`: User's email address (unique)
- `plan`: Current subscription plan (starter or pro)
- `paypal_customer_id`: Reference to PayPal customer
- `subscription_id`: Reference to PayPal subscription
- `subscription_status`: Current subscription status
- `trial_start_date` and `trial_end_date`: For tracking the 7-day free trial period
- `current_period_end`: When the current subscription period ends
- `billing_cycle`: Whether the user is on monthly or yearly billing

**Usage:**
- Authentication: Mapping Clerk authentication to internal users
- Authorization: Checking subscription plans for feature access
- Subscription management: Tracking trial and billing status

### 2. Instances Table

Stores information about n8n instances created by users.

```sql
CREATE TABLE instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    container_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'running', 'stopped', 'error'
    host VARCHAR(255),
    port INTEGER,
    url VARCHAR(255),
    cpu_limit FLOAT, -- CPU cores
    memory_limit INTEGER, -- MB
    storage_limit INTEGER, -- GB
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

**Key Fields:**
- `id`: Unique identifier for the instance
- `user_id`: Owner of the instance
- `name`: Display name for the instance
- `description`: Optional description of the instance
- `container_id`: Docker container ID
- `status`: Current operational status
- `host`: Subdomain part of the URL
- `url`: Full URL for accessing the instance
- `port`: Port number mapped to the container
- `cpu_limit`, `memory_limit`, `storage_limit`: Resource allocations based on plan

**Usage:**
- Container management: Mapping between database records and Docker containers
- URL generation: Building URLs from the subdomain
- Resource management: Enforcing plan-based resource limits
- Status tracking: Maintaining instance lifecycle state

### 3. Resource Usage Table

Tracks resource consumption metrics for each instance.

```sql
CREATE TABLE resource_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES instances(id),
    timestamp TIMESTAMP,
    cpu_usage FLOAT, -- percentage (0-100)
    memory_usage BIGINT, -- in bytes
    disk_usage BIGINT, -- in bytes
    network_in BIGINT, -- in bytes
    network_out BIGINT, -- in bytes
    created_at TIMESTAMP
);
```

**Key Fields:**
- `instance_id`: The instance these metrics belong to
- `timestamp`: When the metrics were recorded
- `cpu_usage`: CPU utilization percentage (0-100)
- `memory_usage`: Memory consumption in bytes
- `disk_usage`: Disk space used in bytes
- `network_in/out`: Network traffic in bytes
- `created_at`: When the record was created

**Usage:**
- Monitoring: Tracking resource usage over time
- Billing: Usage-based billing calculations (future feature)
- Dashboard: Displaying performance metrics to users
- Alerting: Detecting resource constraints or issues

### 4. Payments Table

Records payment transactions and subscription details.

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    amount INTEGER, -- in cents
    currency VARCHAR(3) DEFAULT 'usd',
    status VARCHAR(20), -- 'pending', 'succeeded', 'failed', 'refunded'
    paypal_payment_id VARCHAR(255),
    paypal_order_id VARCHAR(255),
    invoice_url VARCHAR(255),
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**Key Fields:**
- `user_id`: User who made the payment
- `amount`: Payment amount in cents (e.g., 2900 for $29.00)
- `currency`: Payment currency (lowercase, e.g., 'usd')
- `status`: Payment processing status
- `paypal_payment_id`: External ID from PayPal for the payment
- `paypal_order_id`: External ID from PayPal for the order
- `invoice_url`: URL to the hosted invoice
- `description`: Human-readable description of the payment
- `metadata`: Additional payment data in JSON format

**Usage:**
- Subscription management: Tracking payment history
- Billing: Maintaining financial records
- Plan upgrades/downgrades: Recording plan changes
- Receipt generation: Providing payment receipts to users

## Relationships Between Tables

- **User → Instances**: One-to-many relationship. A user can have multiple instances.
- **Instance → Resource Usage**: One-to-many relationship. An instance has multiple resource usage records over time.
- **User → Payments**: One-to-many relationship. A user can have multiple payment records.

## Subscription Plans and Resource Limits

The system supports two subscription tiers with different resource allocations:

| Plan    | Pricing               | Max Instances | CPU/Instance | Memory/Instance | Storage/Instance |
|---------|------------------------|---------------|--------------|-----------------|------------------|
| Starter | $2/mo or $20/yr       | 2             | 1.0 cores    | 1024 MB         | 5 GB             |
| Pro     | $5/mo or $50/yr       | 10            | 2.0 cores    | 2048 MB         | 20 GB            |

The Starter plan includes a 7-day free trial period. After the trial ends, users are billed according to their selected billing cycle (monthly or yearly).

## Data Security Considerations

- **Personal Information**: Only essential user information is stored (email, name)
- **Authentication**: No passwords are stored; authentication is delegated to Clerk
- **Payment Information**: No credit card data is stored; payment processing is handled by PayPal
- **Container Isolation**: Each user's n8n instances are isolated in their own Docker containers
- **Host Bind Mounts**: Data is persisted in host-mounted volumes with proper permissions

## Development Mode

When running in development mode with `DISABLE_PAYMENTS=true`:

1. Authentication is bypassed with a development user (Pro plan)
2. PayPal integration is disabled
3. Clerk verification is skipped
4. All API endpoints remain accessible
5. No trial period limitations are enforced

## Database Migrations

Migrations are handled automatically by GORM at application startup. The `db.RunMigrations()` function:

1. Ensures proper database connection
2. Auto-migrates the tables based on the model definitions
3. Preserves existing data during schema updates

Custom migrations for PayPal integration are provided in the `db/migrations/` directory.

## Example Queries

### Get active instances for a user
```sql
SELECT * FROM instances 
WHERE user_id = 'user_uuid' 
AND status IN ('running', 'stopped') 
AND deleted_at IS NULL;
```

### Get resource usage for an instance (last 24 hours)
```sql
SELECT * FROM resource_usages 
WHERE instance_id = 'instance_uuid' 
AND timestamp > NOW() - INTERVAL '24 hours' 
ORDER BY timestamp DESC;
```

### Check if user has reached their instance limit
```sql
SELECT COUNT(*) FROM instances 
WHERE user_id = 'user_uuid' 
AND deleted_at IS NULL;
```

### Get payment history for a user
```sql
SELECT * FROM payments 
WHERE user_id = 'user_uuid' 
ORDER BY created_at DESC;
```

### Get PayPal payment details
```sql
SELECT * FROM payments
WHERE paypal_order_id = 'order_id';
``` 