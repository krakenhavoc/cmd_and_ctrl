package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const testKey = "0123456789abcdef0123456789abcdef-test-only"

func newTestHMAC(t *testing.T, key string) *HMACAuthenticator {
	t.Helper()
	a, err := NewHMACAuthenticator([]byte(key))
	if err != nil {
		t.Fatalf("NewHMACAuthenticator: %v", err)
	}
	return a
}

// fullPrincipal sets every field of Principal to a non-zero value, by
// reflection, so a field added to Principal without a matching claim
// in hmac.go fails TestHMACRoundTripsEveryField instead of silently
// coming back zero after a restart.
func fullPrincipal(t *testing.T) Principal {
	t.Helper()
	var p Principal
	v := reflect.ValueOf(&p).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		name := v.Type().Field(i).Name
		switch f.Interface().(type) {
		case Role:
			f.Set(reflect.ValueOf(RolePlayer))
		case string:
			f.SetString(name + "-value ü")
		case uuid.UUID:
			f.Set(reflect.ValueOf(uuid.New()))
		case time.Time:
			// Stamped by Issue.
		default:
			t.Fatalf("Principal.%s has type %s, which fullPrincipal does not know how to fill; teach it, and add the field to hmacClaims", name, f.Type())
		}
	}
	return p
}

func TestHMACRoundTripsEveryField(t *testing.T) {
	a := newTestHMAC(t, testKey)
	p := fullPrincipal(t)

	tok, issued, err := a.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	back, err := a.Validate(context.Background(), tok)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	// time.Time carries a monotonic reading and a *Location; compare
	// the instants, then the rest with the times zeroed.
	if !back.IssuedAt.Equal(issued.IssuedAt) || !back.ExpiresAt.Equal(issued.ExpiresAt) {
		t.Errorf("timestamps: got %v/%v, want %v/%v", back.IssuedAt, back.ExpiresAt, issued.IssuedAt, issued.ExpiresAt)
	}
	back.IssuedAt, back.ExpiresAt, issued.IssuedAt, issued.ExpiresAt = time.Time{}, time.Time{}, time.Time{}, time.Time{}
	if back != issued {
		t.Errorf("round trip lost data:\n got  %+v\n want %+v", back, issued)
	}
}

// TestHMACUserIDRoundTrips pins ADR 0051 decision 3: UserID rides the
// credential, and a zero UserID (admin, guest) stays zero.
func TestHMACUserIDRoundTrips(t *testing.T) {
	a := newTestHMAC(t, testKey)
	for _, id := range []uuid.UUID{uuid.New(), uuid.Nil} {
		tok, _, err := a.Issue(context.Background(), Principal{Role: RoleIdentified, UserID: id}, time.Hour)
		if err != nil {
			t.Fatalf("Issue: %v", err)
		}
		back, err := a.Validate(context.Background(), tok)
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if back.UserID != id {
			t.Errorf("UserID: got %v, want %v", back.UserID, id)
		}
	}
}

// TestHMACSurvivesAFreshInstance is the deploy: one process issues,
// the next process, started with the same key, validates.
func TestHMACSurvivesAFreshInstance(t *testing.T) {
	before := newTestHMAC(t, testKey)
	p := Principal{Role: RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), Name: "Alice"}
	tok, _, err := before.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	after := newTestHMAC(t, testKey)
	back, err := after.Validate(context.Background(), tok)
	if err != nil {
		t.Fatalf("Validate on a fresh instance: %v", err)
	}
	if back.GameID != p.GameID || back.PlayerID != p.PlayerID {
		t.Errorf("principal: got %+v, want %+v", back, p)
	}
}

func TestHMACRejectsAnotherKey(t *testing.T) {
	a := newTestHMAC(t, testKey)
	b := newTestHMAC(t, testKey+"-rotated")
	tok, _, err := a.Issue(context.Background(), Principal{Role: RoleAdmin}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := b.Validate(context.Background(), tok); err != ErrInvalidCredential {
		t.Errorf("Validate with another key: got %v, want ErrInvalidCredential", err)
	}
}

func TestHMACRejectsTampering(t *testing.T) {
	a := newTestHMAC(t, testKey)
	p := Principal{Role: RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), Name: "Alice"}
	tok, _, err := a.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("token %q: want 3 dot-separated parts", tok)
	}

	// A meaningful forgery: decode the payload, promote to admin,
	// re-encode, keep the old signature.
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	promoted := strings.Replace(string(payload), `"r":"player"`, `"r":"admin"`, 1)
	if promoted == string(payload) {
		t.Fatalf("payload %s has no role claim to rewrite", payload)
	}
	forged := parts[0] + "." + base64.RawURLEncoding.EncodeToString([]byte(promoted)) + "." + parts[2]

	flip := func(s string, i int) string {
		b := []byte(s)
		if b[i] == 'A' {
			b[i] = 'B'
		} else {
			b[i] = 'A'
		}
		return string(b)
	}

	cases := map[string]string{
		"payload promoted to admin": forged,
		"payload byte flipped":      parts[0] + "." + flip(parts[1], len(parts[1])/2) + "." + parts[2],
		"signature byte flipped":    parts[0] + "." + parts[1] + "." + flip(parts[2], 0),
		"signature truncated":       parts[0] + "." + parts[1] + "." + parts[2][:len(parts[2])-2],
		"signature dropped":         parts[0] + "." + parts[1] + ".",
		"signature padded":          tok + "=",
		"version changed":           "v2." + parts[1] + "." + parts[2],
		"extra segment":             tok + ".x",
	}
	for name, bad := range cases {
		if _, err := a.Validate(context.Background(), bad); err != ErrInvalidCredential {
			t.Errorf("%s: got %v, want ErrInvalidCredential", name, err)
		}
	}
}

func TestHMACRejectsExpired(t *testing.T) {
	a := newTestHMAC(t, testKey)
	start := time.Unix(1_700_000_000, 0).UTC()
	now := start
	a.now = func() time.Time { return now }

	tok, issued, err := a.Issue(context.Background(), Principal{Role: RoleAdmin}, time.Minute)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if !issued.ExpiresAt.Equal(start.Add(time.Minute)) {
		t.Fatalf("ExpiresAt: got %v, want %v", issued.ExpiresAt, start.Add(time.Minute))
	}

	now = start.Add(time.Minute) // the last valid instant
	if _, err := a.Validate(context.Background(), tok); err != nil {
		t.Errorf("Validate at expiry: got %v, want nil", err)
	}
	now = start.Add(time.Minute + time.Millisecond)
	if _, err := a.Validate(context.Background(), tok); err != ErrExpiredCredential {
		t.Errorf("Validate past expiry: got %v, want ErrExpiredCredential", err)
	}
	// Stateless: nothing to purge, so it stays "expired", not "unknown".
	if _, err := a.Validate(context.Background(), tok); err != ErrExpiredCredential {
		t.Errorf("second Validate past expiry: got %v, want ErrExpiredCredential", err)
	}
}

// TestHMACRevokeIsAdvisory records ADR 0044 decision 3: Revoke cannot
// withdraw a stateless token. If this test starts failing because
// Revoke gained teeth (ADR 0051 decision 6), update the type's doc
// comment with it.
func TestHMACRevokeIsAdvisory(t *testing.T) {
	a := newTestHMAC(t, testKey)
	tok, _, err := a.Issue(context.Background(), Principal{Role: RolePlayer}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if err := a.Revoke(context.Background(), tok); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if _, err := a.Validate(context.Background(), tok); err != nil {
		t.Errorf("Validate after advisory Revoke: got %v, want nil", err)
	}
}

// TestHMACTokenIsURLSafeAndBounded: the token rides in a WS ?token=
// and a cookie, so it must need no escaping and stay far below the
// ~4 KiB cookie limit even for the largest principal the lobby mints
// (40-rune name of 4-byte runes, every Discord field at its limit).
func TestHMACTokenIsURLSafeAndBounded(t *testing.T) {
	a := newTestHMAC(t, testKey)
	p := Principal{
		Role:              RolePlayer,
		UserID:            uuid.New(),
		GameID:            uuid.New(),
		PlayerID:          uuid.New(),
		Name:              strings.Repeat("\U0001F600", 40),
		DiscordID:         "123456789012345678901",
		DiscordUsername:   strings.Repeat("u", 32),
		DiscordGlobalName: strings.Repeat("\U0001F600", 32),
		DiscordAvatarHash: "a_" + strings.Repeat("f", 32),
	}
	tok, _, err := a.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if esc := url.QueryEscape(tok); esc != tok {
		t.Errorf("token needs query escaping: %q", tok)
	}
	if len(tok) > 1500 {
		t.Errorf("worst-case token is %d bytes; keep it well under the 4 KiB cookie limit", len(tok))
	}
	t.Logf("worst-case token length: %d", len(tok))

	typical, _, _ := a.Issue(context.Background(), Principal{
		Role: RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), Name: "Alice",
		DiscordID: "123456789012345678", DiscordUsername: "alice", DiscordGlobalName: "Alice",
		DiscordAvatarHash: strings.Repeat("f", 32),
	}, time.Hour)
	t.Logf("typical Discord player token length: %d", len(typical))
}

func TestNewHMACAuthenticatorRefusesShortKey(t *testing.T) {
	for _, k := range []string{"", "short", strings.Repeat("k", MinSessionKeyBytes-1)} {
		if _, err := NewHMACAuthenticator([]byte(k)); !errors.Is(err, ErrSessionKeyTooShort) {
			t.Errorf("key of %d bytes: got %v, want ErrSessionKeyTooShort", len(k), err)
		}
	}
}

type recordingLogger struct{ warnings []string }

func (r *recordingLogger) Warn(msg string, args ...any) {
	r.warnings = append(r.warnings, fmt.Sprint(append([]any{msg}, args...)...))
}

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

// TestNewFromEnvMissingKeyFallsBackLoudly is the acceptance criterion:
// no key means the in-memory store and a warning that names the
// variable, never an HMAC authenticator on some default key.
func TestNewFromEnvMissingKeyFallsBackLoudly(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"unset": {},
		"blank": {SessionKeyEnv: "   "},
	} {
		t.Run(name, func(t *testing.T) {
			log := &recordingLogger{}
			a, err := NewFromEnv(envOf(env), log)
			if err != nil {
				t.Fatalf("NewFromEnv: %v", err)
			}
			if _, ok := a.(*MemoryAuthenticator); !ok {
				t.Errorf("got %T, want *MemoryAuthenticator", a)
			}
			if len(log.warnings) != 1 || !strings.Contains(log.warnings[0], SessionKeyEnv) {
				t.Errorf("want one warning naming %s, got %q", SessionKeyEnv, log.warnings)
			}
		})
	}
}

func TestNewFromEnvShortKeyFailsTheBoot(t *testing.T) {
	log := &recordingLogger{}
	a, err := NewFromEnv(envOf(map[string]string{SessionKeyEnv: "too-short"}), log)
	if !errors.Is(err, ErrSessionKeyTooShort) {
		t.Fatalf("got (%T, %v), want ErrSessionKeyTooShort", a, err)
	}
	if !strings.Contains(err.Error(), SessionKeyEnv) {
		t.Errorf("error %q does not name %s", err, SessionKeyEnv)
	}
}

func TestNewFromEnvWithKeyIsDurable(t *testing.T) {
	log := &recordingLogger{}
	env := envOf(map[string]string{SessionKeyEnv: " " + testKey + "\n"})
	a, err := NewFromEnv(env, log)
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}
	if _, ok := a.(*HMACAuthenticator); !ok {
		t.Fatalf("got %T, want *HMACAuthenticator", a)
	}
	if len(log.warnings) != 0 {
		t.Errorf("unexpected warnings: %q", log.warnings)
	}
	tok, _, err := a.Issue(context.Background(), Principal{Role: RoleAdmin}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	// Whitespace is trimmed, so the untrimmed key's instance agrees.
	if _, err := newTestHMAC(t, testKey).Validate(context.Background(), tok); err != nil {
		t.Errorf("Validate with the trimmed key: %v", err)
	}
}
