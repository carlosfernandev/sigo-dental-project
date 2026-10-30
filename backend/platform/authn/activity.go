package authn

import (
	"errors"
	"sync"
	"time"
)

const IdleTimeout = 15 * time.Minute

var ErrIdleTimeout = errors.New("token idle timeout")

// ActivityTracker keeps last activity for the single-process monolith. A request
// after the timeout is rejected; the token's iat starts the inactivity window.
type ActivityTracker struct {
	mu        sync.Mutex
	lastSeen  map[string]time.Time
	idleLimit time.Duration
	now       func() time.Time
}

func NewActivityTracker() *ActivityTracker {
	return &ActivityTracker{
		lastSeen:  make(map[string]time.Time),
		idleLimit: IdleTimeout,
		now:       time.Now,
	}
}

func (a *ActivityTracker) Touch(tokenID string, issuedAt time.Time) error {
	if tokenID == "" || issuedAt.IsZero() {
		return ErrInvalidToken
	}
	now := a.now().UTC()
	a.mu.Lock()
	defer a.mu.Unlock()

	for id, lastSeen := range a.lastSeen {
		if now.Sub(lastSeen) >= a.idleLimit {
			delete(a.lastSeen, id)
		}
	}

	lastSeen, ok := a.lastSeen[tokenID]
	if !ok {
		lastSeen = issuedAt.UTC()
	}
	if now.Sub(lastSeen) >= a.idleLimit {
		delete(a.lastSeen, tokenID)
		return ErrIdleTimeout
	}
	a.lastSeen[tokenID] = now
	return nil
}
