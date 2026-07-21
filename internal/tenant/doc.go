// Package tenant resolves and carries the tenant a request belongs to.
//
// Every request is attributed to exactly one tenant before it reaches any
// repository: the Resolve middleware determines that tenant from the
// X-Tenant-ID header, or from the configured default in B2C mode,
// validates it against the store, and places its identifier in the request
// context, where FromContext retrieves it.
//
// The identifier is always derived server-side and never read from a request
// body. A request whose tenant cannot be established is rejected rather than
// served without one.
package tenant
