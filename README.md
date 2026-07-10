# Gauthly

**A secure, multi-tenant authentication & authorization API, written in Go.**

Gauthly is a self-hostable backend service that handles everything an application needs to manage its users securely: registration, login, sessions, two-factor authentication, social login, and fine-grained access control — all behind a clean, ready-to-use API. It works out of the box for a single application (B2C) and scales to serving many isolated organizations from one deployment (multi-tenant).

<!-- TODO : badges build / license / go version — ex. shields.io -->
<!-- ![Build](...) ![License](...) ![Go](...) -->

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Security](#security)
- [Roadmap](#roadmap)
- [License](#license)

---

## Overview

Most applications need the same thing: a safe way to know *who* a user is and *what* they are allowed to do. Gauthly provides that as a standalone API, so an application can delegate its entire authentication and authorization layer to a single, well-defined service.

It is built around three ideas:

- **Secure by default** — sane cryptographic choices, protection against common attacks, and strict data isolation are built in, not optional add-ons.
- **Multi-tenant first** — a single deployment can serve one application or thousands of isolated organizations, each with its own users, roles, and data. Single-app usage is simply the case of one tenant.
- **Clean and self-hostable** — a pure API with no vendor lock-in, deployable in one command.

---

## Features

**Authentication**

- Email & password sign-up and login (argon2id password hashing)
- Session management with short-lived access tokens and rotating refresh tokens
- Automatic detection of stolen/replayed tokens
- Two-factor authentication (TOTP) with recovery codes
- Social login via Google and GitHub (OAuth 2.0 + PKCE)
- Password reset and email verification flows

**Authorization**

- Role-based access control (RBAC) with fine-grained permissions
- Permissions scoped per tenant — a role in one organization never leaks into another

**Multi-tenancy**

- Strict data isolation between tenants
- Works identically for single-app (B2C) and multi-organization (B2B / B2B2C) use
- Same email can exist independently across different tenants

**Security & operations**

- Rate limiting on sensitive endpoints
- CSRF protection, configurable CORS, and hardened HTTP headers
- Secrets managed through environment variables
- Auto-generated interactive API documentation (OpenAPI / Swagger)

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26+ |
| HTTP router | chi · net/http |
| Database | PostgreSQL |
| Migrations | goose |
| Cryptography | argon2id · AES-256-GCM · Ed25519 |
| Containerization | Docker & Docker Compose |
| API docs | OpenAPI / Swagger |

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/) 1.26 or later
- [Docker](https://www.docker.com/) & Docker Compose
- [goose](https://github.com/pressly/goose) (database migrations)

### Installation

```bash
# 1. Clone the repository
git clone https://github.com/Harkaso/Gauthly.git
cd Gauthly

# 2. Configure environment variables
cp .env.example .env
# then edit .env with your own values (database URL, secret keys, OAuth credentials…)

# 3. Start the database and services
docker compose up -d

# 4. Apply database migrations
goose -dir ./migrations postgres "$DB_URL" up

# 5. Run the API
go run ./cmd/api
```

The API is now available at `http://localhost:8080`.
Interactive API documentation is served at `http://localhost:8080/docs`.

---

## Usage

Below are a few common requests. All examples use the `X-Tenant-ID` header to identify the tenant (use the default tenant for single-app usage).

**Create an account**

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: default" \
  -d '{"email": "alice@example.com", "password": "a-strong-password"}'
```

**Log in**

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: default" \
  -d '{"email": "alice@example.com", "password": "a-strong-password"}'
```

**Access a protected endpoint**

```bash
curl http://localhost:8080/me \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: default"
```

<!-- TODO : JSON responses examples -->
<!-- TODO : Swagger UI screenshot in /assets -->

---

## API Reference

The full, interactive API reference is auto-generated and available at `/docs` once the service is running.

Main endpoint groups:

| Group | Purpose |
|---|---|
| `/auth` | Registration, login, logout, token refresh |
| `/auth/2fa` | TOTP enrolment and verification <!-- cercle 2 --> |
| `/auth/oauth` | Social login (Google, GitHub) <!-- cercle 2 --> |
| `/users` | User management (permission-gated) |
| `/roles` | Role and permission management (permission-gated) |
| `/tenants` | Tenant administration |

<!-- TODO : adjust exact paths after implementation -->

---

## Project Structure

```
cmd/api/            Application entry point
internal/
  auth/             Authentication logic (register, login, sessions)
  tenant/           Multi-tenancy resolution and isolation
  user/             User domain
  rbac/             Role-based access control
  server/           HTTP setup, routing, middleware
  platform/
    database/       PostgreSQL access
    crypto/         Cryptographic helpers
migrations/         Database migrations (goose)
docs/               Documentation and OpenAPI specification
```

---

## Testing

```bash
# Run the full test suite
go test ./...

# Run a specific package
go test ./internal/auth

# Lint and vulnerability scan
golangci-lint run
govulncheck ./...
```

The test suite includes **offensive security tests** that actively attempt to break the system — cross-tenant data access, token replay, forged tokens, privilege escalation, and more — and verify that each attempt fails.

---

## Security

Security is the core focus of this project. Key measures include:

- **Password storage** with argon2id.
- **Session security** via short-lived Ed25519-signed access tokens and rotating, revocable refresh tokens with theft detection.
- **Strict tenant isolation** enforced at the data-access layer.
- **Deny-by-default authorization** — access is refused unless explicitly granted.
- **Protection** against user enumeration, brute-force, CSRF, and injection.

The project follows the OWASP Application Security Verification Standard (ASVS) and OWASP Top 10 as reference frameworks.

> If you discover a security issue, please open an issue or contact the maintainer rather than disclosing it publicly.

---

## Roadmap

Planned and possible future work:

- WebAuthn / passkeys support
- Passwordless login (magic links / email OTP)
- Attribute-based access control (ABAC), via the existing extensible authorization interface
- Redis-backed sessions and distributed rate limiting for large-scale deployments

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
