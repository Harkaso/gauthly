// Package auth implements the authentication use cases, starting with user
// registration.
//
// AuthService holds the domain logic and depends only on ports, such as
// UserRepository, never on HTTP or SQL. Handler adapts those use cases to HTTP,
// while the repository ports are implemented in the database layer. Every
// operation is scoped to the tenant carried by the request context and never
// taken from client input; a request without a tenant is rejected rather than
// served.
package auth
