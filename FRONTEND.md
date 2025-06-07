# LaunchStack - n8n Hosting Service

LaunchStack is a professional n8n hosting service that offers dedicated resources, unlimited workflows, and 99.9% uptime guarantee at affordable prices starting at just $2/month.

![LaunchStack](public/images/preview/preview.webp)

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Environment Setup](#environment-setup)
- [Development](#development)
- [Authentication](#authentication)
- [Deployment](#deployment)
- [Docker Deployment](#docker-deployment)
- [Server Requirements](#server-requirements)

## Overview

LaunchStack provides a reliable and secure hosting solution for n8n workflows. The platform is designed to be user-friendly, cost-effective, and highly scalable, catering to individual users and businesses of all sizes. The application features user authentication, contact forms, and a modern responsive design.

## Features

- **User Authentication**: Secure sign-in/sign-up with Clerk
- **Affordable Pricing**: Plans starting at just $2/month with 7-day free trial
- **Automatic Updates**: Stay on the latest version of n8n
- **Dedicated Support**: Priority support from n8n experts
- **Enhanced Security**: Enterprise-grade security and regular backups
- **Custom Domain**: Use your own domain for your n8n instance
- **Scalable Resources**: Scale resources based on workflow demands
- **Contact Forms**: Integrated contact forms with Formspree
- **Responsive Design**: Beautiful, modern UI built with Tailwind CSS

## Tech Stack

- **Framework**: Next.js 15.3.3 (App Router)
- **Runtime**: React 18.3.1
- **Language**: TypeScript
- **Authentication**: Clerk
- **Styling**: Tailwind CSS
- **UI Components**: shadcn/ui
- **Animations**: Framer Motion
- **Icons**: Lucide React
- **Form Handling**: React Hook Form with Zod validation
- **Form Backend**: Formspree
- **Fonts**: Montserrat (headings), Work Sans (body)
- **Deployment**: Docker with Alpine Node.js

## Project Structure

```
LaunchStack/
├── app/                     # Next.js App Router pages
│   ├── page.tsx             # Home page
│   ├── layout.tsx           # Root layout with Clerk integration
│   ├── about/               # About page
│   ├── contact/             # Contact page
│   ├── features/            # Features page
│   ├── pricing/             # Pricing page
│   ├── privacy/             # Privacy policy
│   ├── security/            # Security page
│   └── terms/               # Terms page
├── components/              # React components
│   ├── ui/                  # UI components from shadcn/ui
│   ├── header.tsx           # Site header with authentication
│   ├── footer.tsx           # Site footer
│   ├── schema-markup.tsx    # Client-side schema markup
│   └── scroll-button-wrapper.tsx # Client wrapper for scroll button
├── public/                  # Static assets
│   └── images/              # Image assets
├── lib/                     # Utility functions and constants
├── hooks/                   # Custom React hooks
├── middleware.ts            # Clerk authentication middleware
├── .env                     # Environment variables (not in repo)
├── .env.example             # Environment variables template
├── Dockerfile               # Docker configuration
├── docker-compose.yml       # Docker Compose configuration
├── .dockerignore            # Docker ignore file
├── next.config.js           # Next.js configuration
└── tailwind.config.ts       # Tailwind CSS configuration
```

## Environment Setup

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Clerk Authentication
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=your_clerk_publishable_key_here
CLERK_SECRET_KEY=your_clerk_secret_key_here

# Form Handling
NEXT_PUBLIC_FORMSPREE_ENDPOINT=https://formspree.io/f/your_form_id
```

### Getting Clerk Keys

1. Sign up at [Clerk.com](https://clerk.com)
2. Create a new application
3. Copy your publishable key and secret key from the dashboard
4. Add them to your `.env` file

### Setting up Formspree

1. Sign up at [Formspree.io](https://formspree.io)
2. Create a new form
3. Copy the form endpoint URL
4. Add it to your `.env` file as `NEXT_PUBLIC_FORMSPREE_ENDPOINT`

## Development

### Prerequisites

- Node.js 18 or later
- npm or yarn

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/LaunchStack.git
   cd LaunchStack
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your actual values
   ```

4. Start the development server:
   ```bash
   npm run dev
   ```

5. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Authentication

LaunchStack uses Clerk for user authentication, providing:

- **Sign In/Sign Up**: Secure authentication flows
- **User Management**: User profiles and session management
- **Protected Routes**: Middleware-based route protection
- **Social Logins**: Support for various OAuth providers

### Authentication Flow

- Public routes: `/`, `/about`, `/pricing`, `/contact`, `/features`, `/security`, `/privacy`, `/terms`
- Protected routes: All other routes require authentication
- Users can sign in/up from the header navigation
- Authenticated users see their profile in the header

## Deployment

### Build for Production

```bash
npm run build
```

The build output will be in the `launch-stack` directory as configured in `next.config.js`.

### Start Production Server

```bash
npm start
```

## Docker Deployment

LaunchStack can be deployed using Docker for a consistent and isolated environment.

### Environment Variables in Docker

The Docker setup supports environment variables through:

1. `.env` file (for local development)
2. `docker-compose.yml` environment section
3. Docker run command with `-e` flags

### Using Docker

1. Build the Docker image:
   ```bash
   docker build -t launch-stack:latest .
   ```

2. Run the container with environment variables:
   ```bash
   docker run -d -p 3000:3000 \
     -e NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY="your_key" \
     -e CLERK_SECRET_KEY="your_secret" \
     -e NEXT_PUBLIC_FORMSPREE_ENDPOINT="your_endpoint" \
     --name launch-stack launch-stack:latest
   ```

### Using Docker Compose

1. Update environment variables in `docker-compose.yml`:
   ```yaml
   environment:
     - NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=your_clerk_publishable_key
     - CLERK_SECRET_KEY=your_clerk_secret_key
     - NEXT_PUBLIC_FORMSPREE_ENDPOINT=https://formspree.io/f/your_form_id
   ```

2. Start the services:
   ```bash
   docker-compose up -d
   ```

3. View logs:
   ```bash
   docker-compose logs -f
   ```

### Server Proxy Configuration (Caddy)

For production deployment with Caddy as a reverse proxy:

```
launch-stack.srvr.site {
    import tls_config
    
    # Reverse proxy to your Docker container or Node.js server
    reverse_proxy localhost:3000
}
```

## Server Requirements

### Minimum Requirements
- **CPU**: 1 core
- **RAM**: 512MB
- **Storage**: 1GB

### Recommended Requirements
- **CPU**: 2+ cores
- **RAM**: 1GB+
- **Storage**: 5GB+

### Node.js Server
- **Version**: Node.js 18 LTS or later
- **Packages**: npm or yarn
- **Process Manager**: PM2 recommended for production

### Docker Environment
- **Docker**: 20.10.x or later
- **Docker Compose**: 2.x or later

## Notes on Next.js 15 App Router and Server Components

This project uses Next.js 15 App Router with React Server Components. This architecture:

1. Renders components on the server by default
2. Reduces JavaScript sent to the client
3. Improves performance and SEO
4. Requires a Node.js runtime for deployment
5. Uses middleware for authentication routing

Components marked with "use client" directive at the top run on the client side, while all other components run on the server by default.

```tsx
// Client component example
"use client";

import { useState } from "react";

export default function ClientComponent() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}
```

### Key Architectural Patterns

- **Authentication Middleware**: Uses Clerk's `authMiddleware` for route protection
- **Client Component Wrappers**: Components requiring client-side features are wrapped appropriately
- **Environment Variables**: `NEXT_PUBLIC_` variables are available during build time
- **Form Integration**: Server-side form submission with Formspree backend
- **Dynamic Imports**: Used for client-side only components with proper error handling

### Build Optimizations

- Static exports enabled for better performance
- Custom fonts (Montserrat, Work Sans) loaded via next/font
- Image optimization with Next.js Image component
- Chunked JavaScript for better loading performance