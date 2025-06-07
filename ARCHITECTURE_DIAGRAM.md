# LaunchStack - Complete Architecture Diagram

## High-Level System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                  LAUNCHSTACK PAAS                                  │
│                           n8n Hosting Service Architecture                         │
└─────────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                 FRONTEND LAYER                                     │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  Next.js 15 App Router (https://launchstack.io)                                   │
│                                                                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ Landing     │  │ Pricing     │  │ Dashboard   │  │ Instance    │              │
│  │ Page        │  │ Page        │  │ Control     │  │ Management  │              │
│  │ /           │  │ /pricing    │  │ Panel       │  │ /instances  │              │
│  │             │  │             │  │ /dashboard  │  │             │              │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘              │
│        │                 │                 │                 │                   │
│        └─────────────────┼─────────────────┼─────────────────┘                   │
│                          │                 │                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ Contact     │  │ Features    │  │ Security    │  │ Terms/      │              │
│  │ Form        │  │ Page        │  │ Page        │  │ Privacy     │              │
│  │ /contact    │  │ /features   │  │ /security   │  │ /terms      │              │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘              │
│        │                                                                          │
│        └─────────── Formspree ──────────────┐                                    │
└─────────────────────────────────────────────│────────────────────────────────────┘
                                               │
┌─────────────────────────────────────────────│────────────────────────────────────┐
│                            AUTHENTICATION LAYER                                   │
├─────────────────────────────────────────────│────────────────────────────────────┤
│                                             │                                    │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │                        CLERK AUTHENTICATION                                 │ │
│  │                                                                             │ │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │ │
│  │  │ Sign Up/    │    │ JWT Token   │    │ User        │                     │ │
│  │  │ Sign In     │    │ Management  │    │ Management  │                     │ │
│  │  │ Flow        │    │ & Validation│    │ & Sessions  │                     │ │
│  │  └─────────────┘    └─────────────┘    └─────────────┘                     │ │
│  │                                                                             │ │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │ │
│  │  │ Middleware  │    │ Protected   │    │ Webhook     │                     │ │
│  │  │ Auth Guard  │    │ Routes      │    │ Events      │                     │ │
│  │  │ Next.js     │    │ Handler     │    │ Handler     │                     │ │
│  │  └─────────────┘    └─────────────┘    └─────────────┘                     │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────┬───────────────────────────────────────────────┘
                                    │ JWT Bearer Tokens
                                    │ Authorization: Bearer <token>
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                               BACKEND API LAYER                                    │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  Go/Gin Framework API Server (https://api.launchstack.io)                         │
│                                                                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                            MIDDLEWARE STACK                                 │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐     │   │
│  │  │   CORS    │→│   Rate    │→│    JWT    │→│  Request  │→│   Error   │     │   │
│  │  │  Handler  │ │  Limiter  │ │ Validator │ │  Logger   │ │  Handler  │     │   │
│  │  └───────────┘ └───────────┘ └───────────┘ └───────────┘ └───────────┘     │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                           │
│                                        ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                           API ENDPOINTS                                     │   │
│  │                                                                             │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │   │
│  │  │ Health      │  │ User        │  │ Instance    │  │ Metrics     │       │   │
│  │  │ /api/health │  │ /api/user/* │  │ /api/       │  │ /api/       │       │   │
│  │  │             │  │             │  │ instances/* │  │ instances/  │       │   │
│  │  │ GET         │  │ GET/PUT     │  │ CRUD        │  │ {id}/metrics│       │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘       │   │
│  │                                                                             │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │   │
│  │  │ Subscription│  │ Pricing     │  │ Webhooks    │  │ Dashboard   │       │   │
│  │  │ /api/       │  │ /api/       │  │ /api/       │  │ /api/       │       │   │
│  │  │ subscription│  │ pricing/*   │  │ webhooks/*  │  │ dashboard   │       │   │
│  │  │ POST/DELETE │  │ GET         │  │ POST        │  │ GET         │       │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘       │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                           │
│                                        ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                         BUSINESS LOGIC LAYER                               │   │
│  │                                                                             │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │   │
│  │  │ User        │  │ Instance    │  │ Container   │  │ Payment     │       │   │
│  │  │ Service     │  │ Service     │  │ Service     │  │ Service     │       │   │
│  │  │             │  │             │  │             │  │             │       │   │
│  │  │ • CRUD      │  │ • CRUD      │  │ • Docker    │  │ • Stripe    │       │   │
│  │  │ • Profile   │  │ • Lifecycle │  │ • Management│  │ • Billing   │       │   │
│  │  │ • Auth      │  │ • Monitoring│  │ • Networking│  │ • Webhooks  │       │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘       │   │
│  │                                                                             │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │   │
│  │  │ Metrics     │  │ Logging     │  │ Notification│  │ Security    │       │   │
│  │  │ Service     │  │ Service     │  │ Service     │  │ Service     │       │   │
│  │  │             │  │             │  │             │  │             │       │   │
│  │  │ • Resource  │  │ • Audit     │  │ • Alerts    │  │ • Rate      │       │   │
│  │  │ • Usage     │  │ • Errors    │  │ • Status    │  │ • Limiting  │       │   │
│  │  │ • Analytics │  │ • Debug     │  │ • Email     │  │ • Validation│       │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘       │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                                                             │   │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                        DATA PERSISTENCE                            │   │   │
│  │                                                                     │   │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │   │   │
│  │  │ Docker      │  │ N8N Data    │  │ Backup      │                 │   │   │
│  │  │ Volumes     │  │ Directory   │  │ Strategy    │                 │   │   │
│  │  │             │  │             │  │             │                 │   │   │
│  │  │ • Workflows │  │ /opt/n8n/   │  │ • Daily     │                 │   │   │
│  │  │ • Settings  │  │ data/       │  │ • Automated │                 │   │   │
│  │  │ • Logs      │  │             │  │ • S3/Backup │                 │   │   │
│  │  │ └────────────┘  └─────────────┘  └─────────────┘                 │   │   │
│  │  └─────────────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
└───────────────────────────────┬───────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                         CONTAINER INFRASTRUCTURE LAYER                             │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  Docker Engine & Container Management                                             │
│                                                                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                        DOCKER MANAGEMENT                                    │   │
│  │                                                                             │   │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │   │
│  │  │ Docker      │    │ Container   │    │ Network     │                     │   │
│  │  │ Engine      │    │ Registry    │    │ Management  │                     │   │
│  │  │             │    │             │    │             │                     │   │
│  │  │ • API       │    │ • n8n Image │    │ • Isolation │                     │   │
│  │  │ • Runtime   │    │ • Versions  │    │ • Port Map  │                     │   │
│  │  │ • Stats     │    │ • Security  │    │ • Proxy     │                     │   │
│  │  └─────────────┘    └─────────────┘    └─────────────┘                     │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                           │
│                                        ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                       N8N INSTANCES                                         │   │
│  │                                                                             │   │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐             │   │
│  │  │ User 1 Instance │  │ User 2 Instance │  │ User N Instance │             │   │
│  │  │                 │  │                 │  │                 │             │   │
│  │  │ Container:      │  │ Container:      │  │ Container:      │             │   │
│  │  │ user1-n8n       │  │ user2-n8n       │  │ userN-n8n       │             │   │
│  │  │                 │  │                 │  │                 │             │   │
│  │  │ URL:            │  │ URL:            │  │ URL:            │             │   │
│  │  │ user1-n8n.      │  │ user2-n8n.      │  │ userN-n8n.      │             │   │
│  │  │ launchstack.io  │  │ launchstack.io  │  │ launchstack.io  │             │   │
│  │  │                 │  │                 │  │                 │             │   │
│  │  │ Port: 5001      │  │ Port: 5002      │  │ Port: 500N      │             │   │
│  │  │                 │  │                 │  │                 │             │   │
│  │  │ Resources:      │  │ Resources:      │  │ Resources:      │             │   │
│  │  │ CPU: 0.5-2 core │  │ CPU: 0.5-2 core │  │ CPU: 0.5-2 core │             │   │
│  │  │ RAM: 512-2048MB │  │ RAM: 512-2048MB │  │ RAM: 512-2048MB │             │   │
│  │  │ Storage: 5-20GB │  │ Storage: 5-20GB │  │ Storage: 5-20GB │             │   │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘             │   │
│  │                                                                             │   │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │   │
│  │  │                        DATA PERSISTENCE                            │   │   │
│  │  │                                                                     │   │   │
│  │  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │   │   │
│  │  │  │ Docker      │  │ N8N Data    │  │ Backup      │                 │   │   │
│  │  │  │ Volumes     │  │ Directory   │  │ Strategy    │                 │   │   │
│  │  │  │             │  │             │  │             │                 │   │   │
│  │  │  │ • Workflows │  │ /opt/n8n/   │  │ • Daily     │                 │   │   │
│  │  │  │ • Settings  │  │ data/       │  │ • Automated │                 │   │   │
│  │  │  │ • Logs      │  │             │  │ • S3/Backup │                 │   │   │
│  │  │  └─────────────┘  └─────────────┘  └─────────────┘                 │   │   │
│  │  └─────────────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
└───────────────────────────────┬───────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                           EXTERNAL SERVICES LAYER                                  │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ Clerk       │  │ Stripe      │  │ Formspree   │  │ Docker Hub  │              │
│  │ Auth        │  │ Payments    │  │ Forms       │  │ Registry    │              │
│  │             │  │             │  │             │  │             │              │
│  │ • JWT       │  │ • Billing   │  │ • Contact   │  │ • n8n Image │              │
│  │ • Users     │  │ • Webhooks  │  │ • Support   │  │ • Updates   │              │
│  │ • Sessions  │  │ • Plans     │  │ • Feedback  │  │ • Security  │              │
│  │ • Webhooks  │  │ • Invoices  │  │             │  │             │              │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘              │
│                                                                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ DNS/CDN     │  │ SSL/TLS     │  │ Monitoring  │  │ Backup      │              │
│  │ Provider    │  │ Certs       │  │ Services    │  │ Storage     │              │
│  │             │  │             │  │             │  │             │              │
│  │ • Subdomains│  │ • Let's     │  │ • Uptime    │  │ • S3/Cloud  │              │
│  │ • Routing   │  │   Encrypt   │  │ • Alerts    │  │ • Automated │              │
│  │ • Load Bal  │  │ • Auto      │  │ • Metrics   │  │ • Retention │              │
│  │             │  │   Renewal   │  │             │  │             │              │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘              │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Detailed Data Flow Patterns

### 1. User Registration & Onboarding Flow
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ 1. User     │───▶│ 2. Clerk    │───▶│ 3. Webhook  │───▶│ 4. Database │
│ Signs Up    │    │ Creates     │    │ Triggers    │    │ User Record │
│ on Frontend │    │ User        │    │ Backend     │    │ Created     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                            │
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    │
│ 8. Instance │◀───│ 7. Container│◀───│ 6. Trial    │◀───┘
│ Ready &     │    │ Created &   │    │ Activated   │
│ Accessible  │    │ Started     │    │ First Use   │
└─────────────┘    └─────────────┘    └─────────────┘
                          │
                   ┌─────────────┐
                   │ 5. Pricing  │
                   │ Plan        │
                   │ Selection   │
                   └─────────────┘
```

### 2. Instance Creation Flow
```
Frontend Request ──┐
                   │
┌─────────────────────────────────────────────────────────────────┐
│ BACKEND INSTANCE CREATION PIPELINE                             │
│                                                                 │
│ 1. JWT Validation ──┐                                          │
│                     │                                          │
│ 2. Rate Limiting ───┼──┐                                       │
│                     │  │                                       │
│ 3. Plan Validation ─┼──┼──┐                                    │
│                     │  │  │                                    │
│ 4. Resource Check ──┼──┼──┼──┐                                 │
│                     │  │  │  │                                 │
│ 5. Database Record ─┼──┼──┼──┼──┐                              │
│                     │  │  │  │  │                              │
│ 6. Docker Container ┼──┼──┼──┼──┼──┐                           │
│                     │  │  │  │  │  │                           │
│ 7. DNS Configuration┼──┼──┼──┼──┼──┼──┐                        │
│    (AdGuard)        │  │  │  │  │  │  │                        │
└─────────────────────┼──┼──┼──┼──┼──┼──┼────────────────────────┘
                      │  │  │  │  │  │  │
                      ▼  ▼  ▼  ▼  ▼  ▼  ▼
                 Final Response to Frontend
```

### 3. DNS Resolution Flow for Instance Access
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ 1. User     │───▶│ 2. DNS      │───▶│ 3. AdGuard  │───▶│ 4. Docker   │
│ Accesses    │    │ Request for │    │ DNS Rewrites│    │ Container   │
│ Subdomain   │    │ Subdomain   │    │ Two-Level   │    │ Service     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                            │
                                            ▼
                                     ┌─────────────┐
                                     │ 5. N8N      │
                                     │ Workflow    │
                                     │ Interface   │
                                     └─────────────┘
```

DNS Resolution Process:
1. User browser requests `cute-fox.srvr.site`
2. AdGuard DNS resolves `cute-fox.srvr.site` to `cute-fox.docker`
3. AdGuard DNS resolves `cute-fox.docker` to Container IP (e.g., `10.1.2.15`)
4. Traffic reaches the container's N8N service on port 5678
5. User accesses their workflow interface

### 4. Payment & Subscription Flow
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          PAYMENT PROCESSING FLOW                               │
│                                                                                 │
│  Frontend ──┐                                                                  │
│             │                                                                  │
│    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   │
│    │ 1. User     │───▶│ 2. Backend  │───▶│ 3. Stripe   │───▶│ 4. Payment  │   │
│    │ Selects     │    │ Creates     │    │ Processes   │    │ Confirmed   │   │
│    │ Plan        │    │ Intent      │    │ Payment     │    │             │   │
│    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘   │
│                              │                                       │         │
│    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   │
│    │ 8. Instance │◀───│ 7. Resources│◀───│ 6. Database │◀───│ 5. Webhook  │   │
│    │ Activated   │    │ Allocated   │    │ Updated     │    │ Received    │   │
│    │             │    │             │    │             │    │             │   │
│    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘   │
│                                                                                 │
│  Trial Logic:                                                                  │
│  • 7-day free trial for all plans                                             │
│  • Instance created immediately during trial                                  │
│  • Payment required before trial expiration                                   │
│  • Instance suspended if payment fails                                        │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 5. Security & Authentication Patterns
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                             SECURITY ARCHITECTURE                              │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                            FRONTEND SECURITY                            │   │
│  │                                                                         │   │
│  │  • Clerk Authentication (OAuth, Email/Password)                        │   │
│  │  • JWT Token Storage (HTTP-only cookies)                               │   │
│  │  • HTTPS Enforcement                                                    │   │
│  │  • CSP Headers                                                          │   │
│  │  • XSS Protection                                                       │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                         │
│                                      ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                            BACKEND SECURITY                             │   │
│  │                                                                         │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │   │
│  │  │ JWT         │  │ Rate        │  │ Input       │  │ SQL         │   │   │
│  │  │ Validation  │  │ Limiting    │  │ Validation  │  │ Injection   │   │   │
│  │  │             │  │             │  │             │  │ Prevention  │   │   │
│  │  │ • Signature │  │ • Per User  │  │ • Sanitize  │  │ • Prepared  │   │   │
│  │  │ • Expiry    │  │ • Per IP    │  │ • Validate  │  │   Statements│   │   │
│  │  │ • Claims    │  │ • Per Route │  │ • Type Check│  │ • ORM       │   │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                         │
│                                      ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                         CONTAINER SECURITY                              │   │
│  │                                                                         │   │
│  │  • Network Isolation (Docker Networks)                                 │   │
│  │  • Resource Limits (CPU, Memory, Storage)                              │   │
│  │  • Non-root User Execution                                              │   │
│  │  • Read-only File Systems                                               │   │
│  │  • Port Mapping Restrictions                                            │   │
│  │  • Image Security Scanning                                              │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Resource Allocation & Scaling Patterns

```
PLAN-BASED RESOURCE ALLOCATION:

Starter Plan ($2/month):                   Pro Plan ($29/month):
┌─────────────────────────┐               ┌─────────────────────────┐
│ • Max Instances: 1      │               │ • Max Instances: 5      │
│ • CPU: 0.5 cores        │               │ • CPU: 2 cores          │
│ • Memory: 512 MB        │               │ • Memory: 2048 MB       │
│ • Storage: 5 GB         │               │ • Storage: 20 GB        │
│ • Port Range: 5001-5001 │               │ • Port Range: 5001-5005 │
│ • Basic Support         │               │ • Priority Support      │
│ • SSL Certificate       │               │ • Custom Domain         │
│ • 7-day Trial          │               │ • Enhanced Security     │
└─────────────────────────┘               └─────────────────────────┘

PORT ALLOCATION STRATEGY:
Base Port: 5000
User Instance Ports: 5001, 5002, 5003, ..., 5999
Each user gets sequential ports based on instance creation order
```

This comprehensive architecture diagram shows:

1. **Complete system layers** from frontend to infrastructure
2. **Data flow patterns** for key operations
3. **Security architecture** at each layer
4. **Resource allocation strategies** based on pricing plans
5. **External service integrations** with proper boundaries
6. **Container management patterns** for n8n instances
7. **Monitoring and metrics collection** workflows

The ASCII format makes it easy to understand the relationships between components and how data flows through the system during different operations like user onboarding, instance creation, payment processing, and real-time monitoring. 