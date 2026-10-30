package config

// Auth contains authentication settings loaded once at startup.
type Auth struct {
	JWTSecret []byte
}

func loadAuth(l *loader) Auth {
	return Auth{JWTSecret: []byte(l.optional("SIGO_JWT_SECRET"))}
}
