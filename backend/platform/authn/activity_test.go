package authn

import (
	"errors"
	"testing"
	"time"
)

func TestActivityTrackerEnforcesIdleTimeoutAndRefreshesActivity(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	tracker := NewActivityTracker()
	tracker.now = func() time.Time { return now }
	issuedAt := now

	if err := tracker.Touch("token-1", issuedAt); err != nil {
		t.Fatalf("first Touch() error = %v", err)
	}
	now = now.Add(14 * time.Minute)
	if err := tracker.Touch("token-1", issuedAt); err != nil {
		t.Fatalf("active Touch() error = %v", err)
	}
	now = now.Add(14*time.Minute + time.Second)
	if err := tracker.Touch("token-1", issuedAt); err != nil {
		t.Fatalf("refreshed token Touch() error = %v", err)
	}
	now = now.Add(IdleTimeout)
	if err := tracker.Touch("token-1", issuedAt); !errors.Is(err, ErrIdleTimeout) {
		t.Fatalf("idle Touch() error = %v, want %v", err, ErrIdleTimeout)
	}
}

func TestActivityTrackerRejectsTokenAlreadyIdleAtFirstRequest(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	tracker := NewActivityTracker()
	tracker.now = func() time.Time { return now }
	if err := tracker.Touch("token-1", now.Add(-IdleTimeout)); !errors.Is(err, ErrIdleTimeout) {
		t.Fatalf("Touch() error = %v, want %v", err, ErrIdleTimeout)
	}
}

func TestRequireAnyRole(t *testing.T) {
	principal := Principal{Role: "odontologo"}
	if err := RequireAnyRole(principal, "odontologo", "admin_ti"); err != nil {
		t.Fatalf("RequireAnyRole() error = %v", err)
	}
	if err := RequireAnyRole(principal, "recepcionista_cajero"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("RequireAnyRole() error = %v, want %v", err, ErrForbidden)
	}
}
