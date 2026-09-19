package users

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

const testKey = "0123456789abcdef0123456789abcdef-identity"

func openStore(t *testing.T, sealer *Sealer) (*SQLStore, *db.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return NewSQLStore(d, sealer), d
}

func mustSealer(t *testing.T, key string) *Sealer {
	t.Helper()
	s, err := NewSealer(key)
	if err != nil {
		t.Fatalf("NewSealer: %v", err)
	}
	return s
}

// clock pins the store's time so last_seen_at can be asserted.
func clock(s *SQLStore, at time.Time) { s.now = func() time.Time { return at } }

var alice = discord.User{ID: "111", Username: "alice", GlobalName: "Alice", Avatar: "avhash"}

// rawRefreshToken reads the refresh_token column as stored.
func rawRefreshToken(t *testing.T, d *db.DB, subject string) []byte {
	t.Helper()
	var b []byte
	if err := d.QueryRow(`SELECT refresh_token FROM identities WHERE provider = 'discord' AND subject = ?`, subject).Scan(&b); err != nil {
		t.Fatalf("read refresh_token: %v", err)
	}
	return b
}

func TestUpsertFirstSignInMintsAUser(t *testing.T) {
	s, d := openStore(t, mustSealer(t, testKey))
	t0 := time.UnixMilli(1_700_000_000_123).UTC()
	clock(s, t0)

	u, err := s.UpsertFromDiscord(context.Background(), alice, "rt-1", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	if u.ID == uuid.Nil {
		t.Fatal("first sign-in returned a zero user id")
	}
	if u.DisplayName != "Alice" || u.AvatarURL != "/avatars/111/avhash.png" {
		t.Errorf("user = %+v", u)
	}
	if !u.CreatedAt.Equal(t0) || !u.LastSeenAt.Equal(t0) {
		t.Errorf("timestamps = %v / %v, want %v (ms precision)", u.CreatedAt, u.LastSeenAt, t0)
	}

	var (
		userID, name, scopes string
		linked               int64
	)
	if err := d.QueryRow(`SELECT user_id, display_name, scopes, linked_at FROM identities WHERE provider = 'discord' AND subject = '111'`).
		Scan(&userID, &name, &scopes, &linked); err != nil {
		t.Fatalf("read identity: %v", err)
	}
	if userID != u.ID.String() || name != "Alice" || scopes != "identify" || linked != t0.UnixMilli() {
		t.Errorf("identity = %s %q %q %d", userID, name, scopes, linked)
	}

	got, err := s.Get(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != u {
		t.Errorf("Get = %+v, want %+v", got, u)
	}
}

func TestUpsertLaterSignInUpdatesTheSameUser(t *testing.T) {
	s, d := openStore(t, mustSealer(t, testKey))
	t0 := time.UnixMilli(1_700_000_000_000).UTC()
	clock(s, t0)
	first, err := s.UpsertFromDiscord(context.Background(), alice, "rt-1", "identify")
	if err != nil {
		t.Fatalf("first: %v", err)
	}

	renamed := alice
	renamed.GlobalName = "Alice the Bold"
	renamed.Avatar = "newhash"
	t1 := t0.Add(time.Hour)
	clock(s, t1)
	second, err := s.UpsertFromDiscord(context.Background(), renamed, "rt-2", "identify")
	if err != nil {
		t.Fatalf("second: %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("second sign-in minted a new user %s, want %s", second.ID, first.ID)
	}
	if second.DisplayName != "Alice the Bold" || second.AvatarURL != "/avatars/111/newhash.png" {
		t.Errorf("profile not refreshed: %+v", second)
	}
	if !second.CreatedAt.Equal(t0) || !second.LastSeenAt.Equal(t1) {
		t.Errorf("created %v last_seen %v, want %v / %v", second.CreatedAt, second.LastSeenAt, t0, t1)
	}
	var users, identities int
	_ = d.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&users)
	_ = d.QueryRow(`SELECT COUNT(*) FROM identities`).Scan(&identities)
	if users != 1 || identities != 1 {
		t.Errorf("rows: users=%d identities=%d, want 1/1", users, identities)
	}
	var avatarHash string
	_ = d.QueryRow(`SELECT avatar_hash FROM identities WHERE subject = '111'`).Scan(&avatarHash)
	if avatarHash != "newhash" {
		t.Errorf("identity avatar_hash = %q", avatarHash)
	}
	tok, ok, err := s.RefreshToken(context.Background(), "111")
	if err != nil || !ok || tok != "rt-2" {
		t.Errorf("RefreshToken = %q %v %v, want the newer rt-2", tok, ok, err)
	}

	// A sign-in that returns no refresh token keeps the stored one.
	if _, err := s.UpsertFromDiscord(context.Background(), renamed, "", "identify"); err != nil {
		t.Fatalf("third: %v", err)
	}
	if tok, ok, _ := s.RefreshToken(context.Background(), "111"); !ok || tok != "rt-2" {
		t.Errorf("empty refresh token dropped the stored one: %q %v", tok, ok)
	}
}

func TestIdentityIsUnique(t *testing.T) {
	s, d := openStore(t, nil)
	a, err := s.UpsertFromDiscord(context.Background(), alice, "", "identify")
	if err != nil {
		t.Fatalf("alice: %v", err)
	}
	b, err := s.UpsertFromDiscord(context.Background(), discord.User{ID: "222", Username: "bob"}, "", "identify")
	if err != nil {
		t.Fatalf("bob: %v", err)
	}
	if a.ID == b.ID {
		t.Fatal("two snowflakes share one user")
	}
	if b.DisplayName != "bob" || b.AvatarURL != "" {
		t.Errorf("bob = %+v (username fallback, no avatar)", b)
	}

	// (provider, subject) is the primary key: one snowflake, one row.
	if _, err := d.Exec(`INSERT INTO identities (provider, subject, user_id, display_name, scopes, linked_at) VALUES ('discord', '111', ?, 'x', 'identify', 1)`,
		b.ID.String()); err == nil {
		t.Error("a second identities row for the same Discord account was accepted")
	}
	// (user_id, provider) is unique: one Discord account per user.
	if _, err := d.Exec(`INSERT INTO identities (provider, subject, user_id, display_name, scopes, linked_at) VALUES ('discord', '333', ?, 'x', 'identify', 1)`,
		a.ID.String()); err == nil {
		t.Error("a second Discord account was linked to the same user")
	}
	// ...but the same user may hold an identity at another provider.
	if _, err := d.Exec(`INSERT INTO identities (provider, subject, user_id, display_name, scopes, linked_at) VALUES ('entra', 'oid-1', ?, 'x', 'openid', 1)`,
		a.ID.String()); err != nil {
		t.Errorf("a second provider for the same user was refused: %v", err)
	}
	// An identity must point at a real user.
	if _, err := d.Exec(`INSERT INTO identities (provider, subject, user_id, display_name, scopes, linked_at) VALUES ('discord', '444', ?, 'x', 'identify', 1)`,
		uuid.NewString()); err == nil {
		t.Error("an identity for a user that does not exist was accepted")
	}
}

func TestMissingKeyDiscardsTheRefreshToken(t *testing.T) {
	s, d := openStore(t, nil)
	if _, err := s.UpsertFromDiscord(context.Background(), alice, "rt-secret", "identify"); err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	if b := rawRefreshToken(t, d, "111"); b != nil {
		t.Errorf("refresh_token = %q, want NULL with no key", b)
	}
	if _, ok, err := s.RefreshToken(context.Background(), "111"); ok || err != nil {
		t.Errorf("RefreshToken with nothing stored = ok %v err %v", ok, err)
	}
}

func TestRemovingTheKeyClearsAStoredToken(t *testing.T) {
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer d.Close()
	withKey := NewSQLStore(d, mustSealer(t, testKey))
	if _, err := withKey.UpsertFromDiscord(context.Background(), alice, "rt-1", "identify"); err != nil {
		t.Fatalf("with key: %v", err)
	}
	if rawRefreshToken(t, d, "111") == nil {
		t.Fatal("token not stored with a key configured")
	}
	withoutKey := NewSQLStore(d, nil)
	if _, err := withoutKey.UpsertFromDiscord(context.Background(), alice, "rt-2", "identify"); err != nil {
		t.Fatalf("without key: %v", err)
	}
	if b := rawRefreshToken(t, d, "111"); b != nil {
		t.Errorf("refresh_token = %x after a keyless sign-in, want NULL", b)
	}
}

func TestStoredRefreshTokenIsNotPlaintext(t *testing.T) {
	s, d := openStore(t, mustSealer(t, testKey))
	if _, err := s.UpsertFromDiscord(context.Background(), alice, "rt-plaintext-marker", "identify"); err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	b := rawRefreshToken(t, d, "111")
	if bytes.Contains(b, []byte("rt-plaintext-marker")) {
		t.Fatal("refresh token stored in the clear")
	}
	if len(b) == 0 || b[0] != sealVersion1 {
		t.Errorf("stored token has no version-1 prefix: %x", b)
	}
}

func TestStoredTokenRefusesTheWrongKey(t *testing.T) {
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer d.Close()
	if _, err := NewSQLStore(d, mustSealer(t, testKey)).UpsertFromDiscord(context.Background(), alice, "rt-1", "identify"); err != nil {
		t.Fatalf("seal: %v", err)
	}
	other := NewSQLStore(d, mustSealer(t, strings.Repeat("z", 40)))
	if _, _, err := other.RefreshToken(context.Background(), "111"); !errors.Is(err, ErrSealed) {
		t.Errorf("RefreshToken under another key: err %v, want ErrSealed", err)
	}
}

func TestSealerRoundTrip(t *testing.T) {
	s := mustSealer(t, testKey)
	ad := []byte("discord\x00111")
	a, err := s.Seal([]byte("secret"), ad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	b, _ := s.Seal([]byte("secret"), ad)
	if bytes.Equal(a, b) {
		t.Error("two seals of the same plaintext are identical — the nonce is not random")
	}
	for _, sealed := range [][]byte{a, b} {
		pt, err := s.Open(sealed, ad)
		if err != nil || string(pt) != "secret" {
			t.Errorf("Open = %q %v", pt, err)
		}
	}
}

func TestSealerRefusesTampering(t *testing.T) {
	s := mustSealer(t, testKey)
	ad := []byte("discord\x00111")
	sealed, err := s.Seal([]byte("secret"), ad)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	for i := range sealed {
		bad := append([]byte(nil), sealed...)
		bad[i] ^= 0x01
		if _, err := s.Open(bad, ad); !errors.Is(err, ErrSealed) {
			t.Errorf("flipped byte %d: err %v, want ErrSealed", i, err)
		}
	}
	if _, err := s.Open(sealed[:len(sealed)-1], ad); !errors.Is(err, ErrSealed) {
		t.Errorf("truncated: err %v", err)
	}
	if _, err := s.Open(nil, ad); !errors.Is(err, ErrSealed) {
		t.Errorf("empty: err %v", err)
	}
	// Bound to its row: another identity's associated data fails.
	if _, err := s.Open(sealed, []byte("discord\x00222")); !errors.Is(err, ErrSealed) {
		t.Errorf("other identity's AD: err %v, want ErrSealed", err)
	}
}

func TestSealerRefusesTheWrongKey(t *testing.T) {
	sealed, err := mustSealer(t, testKey).Seal([]byte("secret"), nil)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if _, err := mustSealer(t, testKey+"x").Open(sealed, nil); !errors.Is(err, ErrSealed) {
		t.Errorf("wrong key: err %v, want ErrSealed", err)
	}
}

type recordingLogger struct{ msgs []string }

func (r *recordingLogger) Warn(msg string, _ ...any) { r.msgs = append(r.msgs, msg) }

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestNewSealerFromEnv(t *testing.T) {
	admin := strings.Repeat("a", 40)
	session := strings.Repeat("s", 40)

	t.Run("absent warns and returns nil", func(t *testing.T) {
		log := &recordingLogger{}
		s, err := NewSealerFromEnv(envOf(map[string]string{"CMDCTRL_ADMIN_TOKEN": admin}), log)
		if err != nil || s != nil {
			t.Fatalf("got %v %v, want nil, nil", s, err)
		}
		if len(log.msgs) != 1 || !strings.Contains(log.msgs[0], IdentityKeyEnv) {
			t.Errorf("warning must name %s: %q", IdentityKeyEnv, log.msgs)
		}
	})
	t.Run("too short fails", func(t *testing.T) {
		_, err := NewSealerFromEnv(envOf(map[string]string{IdentityKeyEnv: "short"}), &recordingLogger{})
		if !errors.Is(err, ErrIdentityKeyTooShort) {
			t.Errorf("err %v, want ErrIdentityKeyTooShort", err)
		}
	})
	t.Run("equal to the admin token fails", func(t *testing.T) {
		_, err := NewSealerFromEnv(envOf(map[string]string{IdentityKeyEnv: admin, "CMDCTRL_ADMIN_TOKEN": admin}), &recordingLogger{})
		if !errors.Is(err, ErrIdentityKeyReused) {
			t.Errorf("err %v, want ErrIdentityKeyReused", err)
		}
	})
	t.Run("equal to the session key fails", func(t *testing.T) {
		_, err := NewSealerFromEnv(envOf(map[string]string{IdentityKeyEnv: " " + session + "\n", "CMDCTRL_SESSION_KEY": session}), &recordingLogger{})
		if !errors.Is(err, ErrIdentityKeyReused) {
			t.Errorf("err %v, want ErrIdentityKeyReused", err)
		}
	})
	t.Run("good key", func(t *testing.T) {
		s, err := NewSealerFromEnv(envOf(map[string]string{IdentityKeyEnv: testKey, "CMDCTRL_ADMIN_TOKEN": admin, "CMDCTRL_SESSION_KEY": session}), &recordingLogger{})
		if err != nil || s == nil {
			t.Fatalf("got %v %v", s, err)
		}
	})
}

func TestNoStore(t *testing.T) {
	var s Store = NoStore{}
	u, err := s.UpsertFromDiscord(context.Background(), alice, "rt", "identify")
	if err != nil || u.ID != uuid.Nil {
		t.Errorf("NoStore upsert = %+v %v, want the zero user", u, err)
	}
	if _, err := s.Get(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("NoStore Get err %v", err)
	}
}

func TestGetUnknownUser(t *testing.T) {
	s, _ := openStore(t, nil)
	if _, err := s.Get(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("err %v, want ErrNotFound", err)
	}
}
