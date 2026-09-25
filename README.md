# HNU Tickets

REST API for a technical support ticket system. The service uses a modular monolith architecture, PostgreSQL, JWT authentication, and role-based access control.

## Quick start

1. Copy `.env.example` to `.env` and set a secure `JWT_SECRET`.
2. Start PostgreSQL and the API with `docker compose up --build`.
3. Check `http://localhost:8080/health` and `http://localhost:8080/ready`.

The complete setup, API reference, architecture, and demo flow are documented in `docs/`.
