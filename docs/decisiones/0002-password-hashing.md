# Decision 0002: Password hashing

**Status:** Accepted by repository owner through `AGENTS.md`  
**Scope:** Identity passwords

## Decision

Use `golang.org/x/crypto/bcrypt` for password hashing. This follows the repository's
owner-approved stack; the backlog's Argon2id suggestion does not override it.

- Minimum: 12 Unicode characters.
- Maximum: 72 UTF-8 bytes, imposed by bcrypt.
- Cost: 12. Hashes below cost 12 are replaced after a successful verification.
- Reject a short embedded list of common passwords and any password containing the account's
  normalized email, case-insensitively.
- Passwords are never logged or stored outside their bcrypt hash.

The cost is a versioned parameter in `internal/identity/passwords`. Revisit it when the
production hardware and login latency budget are measured; increase it when the authentication
capacity test allows it. New password values must pass the policy before hashing. Existing
passwords are verified first and rehashed on successful login when their bcrypt cost is lower.
