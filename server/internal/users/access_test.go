package users

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// Tests for ADR 0156 §1's store half: the last access check, and the
// rotated refresh token a renewal writes back.

func TestRecordAccessCheckRoundTrip(t *testing.T) {
	s, _ := openStore(t, nil)
	ctx := context.Background()
	u, err := s.UpsertFromDiscord(ctx, alice, "", "identify")
	if err != nil {
		t.Fatal(err)
	}
	if at, res, err := s.AccessCheck(ctx, u.ID); err != nil || !at.IsZero() || res != "" {
		t.Fatalf("a new user: %v %q %v, want unchecked", at, res, err)
	}
	when := time.UnixMilli(1_760_000_000_123).UTC()
	for _, r := range []discord.Access{discord.AccessAdmitted, discord.AccessNoRole, discord.AccessNotMember} {
		if err := s.RecordAccessCheck(ctx, u.ID, r, when); err != nil {
			t.Fatalf("record %q: %v", r, err)
		}
		at, got, err := s.AccessCheck(ctx, u.ID)
		if err != nil || !at.Equal(when) || got != r {
			t.Errorf("after recording %q: %v %q %v", r, at, got, err)
		}
	}
	if err := s.RecordAccessCheck(ctx, u.ID, "outage", when); !errors.Is(err, ErrInvalidAccess) {
		t.Errorf("an outage recorded: %v, want ErrInvalidAccess", err)
	}
	if err := s.RecordAccessCheck(ctx, uuid.New(), discord.AccessAdmitted, when); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown user: %v, want ErrNotFound", err)
	}
}

func TestSetRefreshTokenRotatesTheSealedToken(t *testing.T) {
	s, d := openStore(t, mustSealer(t, testKey))
	ctx := context.Background()
	if _, err := s.UpsertFromDiscord(ctx, alice, "rt-1", "identify guilds.members.read"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetRefreshToken(ctx, alice.ID, "rt-2"); err != nil {
		t.Fatalf("SetRefreshToken: %v", err)
	}
	if raw := rawRefreshToken(t, d, alice.ID); raw == nil || bytes.Contains(raw, []byte("rt-2")) {
		t.Errorf("column = %q, want sealed", raw)
	}
	if rt, ok, err := s.RefreshToken(ctx, alice.ID); err != nil || !ok || rt != "rt-2" {
		t.Errorf("RefreshToken = %q %v %v, want rt-2", rt, ok, err)
	}
	// An empty token leaves the working one in place.
	if err := s.SetRefreshToken(ctx, alice.ID, ""); err != nil {
		t.Fatal(err)
	}
	if rt, _, _ := s.RefreshToken(ctx, alice.ID); rt != "rt-2" {
		t.Errorf("an empty rotation erased the token: %q", rt)
	}
	if err := s.SetRefreshToken(ctx, "999", "rt-x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown identity: %v, want ErrNotFound", err)
	}
}

func TestSetRefreshTokenWithNoKeyStoresNothing(t *testing.T) {
	s, d := openStore(t, nil)
	ctx := context.Background()
	if _, err := s.UpsertFromDiscord(ctx, alice, "rt-1", "identify"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetRefreshToken(ctx, alice.ID, "rt-2"); err != nil {
		t.Fatal(err)
	}
	if raw := rawRefreshToken(t, d, alice.ID); raw != nil {
		t.Errorf("with no key the column holds %q, want NULL", raw)
	}
}
