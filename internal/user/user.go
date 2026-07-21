package user

import "github.com/google/uuid"

// User is an account that can authenticate within a single tenant.
//
// Email is unique per tenant rather than globally: the same address may exist
// under different tenants, enforced by the UNIQUE(email, tenant_id) constraint.
// PasswordHash holds the argon2id PHC string produced by crypto.HashPassword,
// never a plaintext password. User is a domain entity, not a response payload:
// it is never serialized directly to a client, so that PasswordHash cannot leak.
type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        string
	PasswordHash string
}
