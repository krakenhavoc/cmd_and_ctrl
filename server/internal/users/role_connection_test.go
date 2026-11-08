package users

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

// Tests for ADR 0163 §4's store half: the role-connection grant, sealed
// in its own column, apart from the sign-in grant.

func rawRoleConnection(t *testing.T, s *SQLStore, subject string) []byte {
	t.Helper()
	var b []byte
	if err := s.db.QueryRow(`SELECT role_connection_token FROM identities WHERE provider = 'discord' AND subject = ?`, subject).Scan(&b); err != nil {
		t.Fatalf("read role_connection_token: %v", err)
	}
	return b
}

func TestRoleConnectionTokenRoundTripIsSealedAndApartFromSignIn(t *testing.T) {
	s, _ := openStore(t, mustSealer(t, testKey))
	ctx := context.Background()
	if _, err := s.UpsertFromDiscord(ctx, alice, "signin-1", "identify"); err != nil {
		t.Fatal(err)
	}
	if !s.CanSeal() {
		t.Fatal("CanSeal false with a key")
	}
	if rt, ok, err := s.RoleConnectionToken(ctx, alice.ID); err != nil || ok || rt != "" {
		t.Fatalf("a new identity: %q %v %v, want not linked", rt, ok, err)
	}
	if err := s.SetRoleConnectionToken(ctx, alice.ID, "rc-1"); err != nil {
		t.Fatal(err)
	}
	raw := rawRoleConnection(t, s, alice.ID)
	if raw == nil || bytes.Contains(raw, []byte("rc-1")) {
		t.Errorf("column = %q, want sealed", raw)
	}
	if rt, ok, err := s.RoleConnectionToken(ctx, alice.ID); err != nil || !ok || rt != "rc-1" {
		t.Errorf("RoleConnectionToken = %q %v %v, want rc-1", rt, ok, err)
	}
	// The sealed blob opens only as this column of this identity.
	if _, err := s.sealer.Open(raw, refreshTokenAD(alice.ID)); err == nil {
		t.Error("the role-connection blob opens as a sign-in refresh token")
	}
	if _, err := s.OpenRoleConnectionToken("222", raw); err == nil {
		t.Error("the blob opens for another identity")
	}
	if rt, err := s.OpenRoleConnectionToken(alice.ID, raw); err != nil || rt != "rc-1" {
		t.Errorf("OpenRoleConnectionToken = %q %v", rt, err)
	}

	// A sign-in replaces its own grant and leaves this one alone.
	if _, err := s.UpsertFromDiscord(ctx, alice, "signin-2", "identify"); err != nil {
		t.Fatal(err)
	}
	if rt, _, _ := s.RoleConnectionToken(ctx, alice.ID); rt != "rc-1" {
		t.Errorf("after a sign-in the role-connection token is %q, want rc-1", rt)
	}
	if rt, _, _ := s.RefreshToken(ctx, alice.ID); rt != "signin-2" {
		t.Errorf("sign-in token = %q", rt)
	}

	// Rotation; an empty token keeps the working one.
	if err := s.SetRoleConnectionToken(ctx, alice.ID, "rc-2"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetRoleConnectionToken(ctx, alice.ID, ""); err != nil {
		t.Fatal(err)
	}
	if rt, _, _ := s.RoleConnectionToken(ctx, alice.ID); rt != "rc-2" {
		t.Errorf("after rotation = %q, want rc-2", rt)
	}

	if err := s.ClearRoleConnectionToken(ctx, alice.ID); err != nil {
		t.Fatal(err)
	}
	if raw := rawRoleConnection(t, s, alice.ID); raw != nil {
		t.Errorf("cleared column = %x, want NULL", raw)
	}
	if rt, _, _ := s.RefreshToken(ctx, alice.ID); rt != "signin-2" {
		t.Errorf("clearing the role connection touched the sign-in token: %q", rt)
	}

	if err := s.SetRoleConnectionToken(ctx, "999", "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown identity: %v, want ErrNotFound", err)
	}
	if _, _, err := s.RoleConnectionToken(ctx, "999"); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown identity read: %v, want ErrNotFound", err)
	}
}

func TestRoleConnectionTokenWithNoKeyIsRefused(t *testing.T) {
	s, _ := openStore(t, nil)
	ctx := context.Background()
	if _, err := s.UpsertFromDiscord(ctx, alice, "", "identify"); err != nil {
		t.Fatal(err)
	}
	if s.CanSeal() {
		t.Error("CanSeal true with no key")
	}
	if err := s.SetRoleConnectionToken(ctx, alice.ID, "rc"); !errors.Is(err, ErrNoSealer) {
		t.Errorf("SetRoleConnectionToken with no key: %v, want ErrNoSealer", err)
	}
	if raw := rawRoleConnection(t, s, alice.ID); raw != nil {
		t.Errorf("with no key the column holds %x", raw)
	}
}
