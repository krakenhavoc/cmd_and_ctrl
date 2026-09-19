package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mapList is a RevocationList over a plain map, with the same "at or
// before" rule users.Revocations uses.
type mapList map[uuid.UUID]time.Time

func (m mapList) Revoked(id uuid.UUID, issuedAt time.Time) bool {
	w, ok := m[id]
	return ok && !issuedAt.After(w)
}

func TestWithRevocationRefusesARevokedUsersOldToken(t *testing.T) {
	ctx := context.Background()
	h := newTestHMAC(t, testKey)
	start := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	now := start
	h.now = func() time.Time { return now }

	user := uuid.New()
	list := mapList{}
	a := WithRevocation(h, list)

	old, _, err := a.Issue(ctx, Principal{Role: RoleIdentified, UserID: user}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := a.Validate(ctx, old); err != nil {
		t.Fatalf("before revocation: %v", err)
	}

	// The user signs out everywhere a minute later.
	now = start.Add(time.Minute)
	list[user] = now
	if _, err := a.Validate(ctx, old); !errors.Is(err, ErrRevokedCredential) {
		t.Fatalf("old token after revocation: err = %v, want ErrRevokedCredential", err)
	}

	// A token issued in the same millisecond as the revocation is
	// covered too: "at or before", so the request that revoked can
	// never race its own watermark.
	same, _, _ := a.Issue(ctx, Principal{Role: RoleIdentified, UserID: user}, time.Hour)
	if _, err := a.Validate(ctx, same); !errors.Is(err, ErrRevokedCredential) {
		t.Errorf("token issued at the watermark: err = %v, want ErrRevokedCredential", err)
	}

	// Signing in again afterwards works.
	now = now.Add(time.Millisecond)
	fresh, _, _ := a.Issue(ctx, Principal{Role: RoleIdentified, UserID: user}, time.Hour)
	if p, err := a.Validate(ctx, fresh); err != nil || p.UserID != user {
		t.Errorf("new token after revocation: %+v, %v", p, err)
	}

	// Someone else's session is untouched.
	other, _, _ := a.Issue(ctx, Principal{Role: RoleIdentified, UserID: uuid.New()}, time.Hour)
	if _, err := a.Validate(ctx, other); err != nil {
		t.Errorf("another user's session: %v", err)
	}
}

func TestWithRevocationIgnoresSessionsWithNoUser(t *testing.T) {
	ctx := context.Background()
	h := newTestHMAC(t, testKey)
	// A list that says "revoked" for everything: a zero-UserID
	// principal must never even be asked about.
	a := WithRevocation(h, alwaysRevoked{t: t})

	for _, p := range []Principal{
		{Role: RoleAdmin, AdminID: uuid.New()},
		{Role: RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), Name: "guest"},
		{Role: RoleSpectator, GameID: uuid.New()},
	} {
		tok, _, err := a.Issue(ctx, p, time.Hour)
		if err != nil {
			t.Fatalf("Issue %s: %v", p.Role, err)
		}
		if _, err := a.Validate(ctx, tok); err != nil {
			t.Errorf("%s session with no UserID: %v", p.Role, err)
		}
	}
}

type alwaysRevoked struct{ t *testing.T }

func (a alwaysRevoked) Revoked(id uuid.UUID, _ time.Time) bool {
	if id == uuid.Nil {
		a.t.Error("RevocationList asked about uuid.Nil")
	}
	return true
}

func TestWithRevocationChecksExpiryFirst(t *testing.T) {
	ctx := context.Background()
	h := newTestHMAC(t, testKey)
	start := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	now := start
	h.now = func() time.Time { return now }
	user := uuid.New()
	a := WithRevocation(h, mapList{user: start.Add(time.Hour)})

	tok, _, _ := a.Issue(ctx, Principal{Role: RoleIdentified, UserID: user}, time.Minute)
	now = start.Add(2 * time.Minute)
	if _, err := a.Validate(ctx, tok); !errors.Is(err, ErrExpiredCredential) {
		t.Errorf("expired and revoked: err = %v, want ErrExpiredCredential", err)
	}
}

func TestWithRevocationNilListIsTheInnerAuthenticator(t *testing.T) {
	h := newTestHMAC(t, testKey)
	if got := WithRevocation(h, nil); got != Authenticator(h) {
		t.Errorf("WithRevocation(h, nil) = %T, want the HMAC authenticator itself", got)
	}
}

func TestMiddlewareSaysRevoked(t *testing.T) {
	ctx := context.Background()
	h := newTestHMAC(t, testKey)
	user := uuid.New()
	list := mapList{}
	a := WithRevocation(h, list)
	tok, issued, _ := a.Issue(ctx, Principal{Role: RoleIdentified, UserID: user}, time.Hour)
	list[user] = issued.IssuedAt

	srv := Middleware(a)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("handler reached with a revoked session")
	}))
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", rec.Code)
	}
	if body := rec.Body.String(); body != "{\"error\":\"session revoked\"}\n" {
		t.Errorf("body %q", body)
	}
}
