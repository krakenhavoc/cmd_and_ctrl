package users

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

var alice2 = discord.User{ID: "222", Username: "bob", GlobalName: "Bob"}

// Tests for users.last_deck (ADR 0110 section 5, migration 0008).

func signedInUser(t *testing.T, s *SQLStore) uuid.UUID {
	t.Helper()
	u, err := s.UpsertFromDiscord(context.Background(), alice, "", "identify")
	if err != nil {
		t.Fatalf("UpsertFromDiscord: %v", err)
	}
	return u.ID
}

func TestLastDeckIsNoneUntilSet(t *testing.T) {
	s, _ := openStore(t, nil)
	id := signedInUser(t, s)
	got, err := s.LastDeck(context.Background(), id)
	if err != nil || !got.IsZero() {
		t.Fatalf("LastDeck = %+v, %v; want none", got, err)
	}
}

func TestSetLastDeckRoundTripsBothKinds(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	id := signedInUser(t, s)

	lib := LastDeck{Kind: LastDeckLibrary, ID: uuid.New().String()}
	if err := s.SetLastDeck(ctx, id, lib); err != nil {
		t.Fatalf("SetLastDeck(library): %v", err)
	}
	if got, err := s.LastDeck(ctx, id); err != nil || got != lib {
		t.Errorf("LastDeck = %+v, %v; want %+v", got, err, lib)
	}
	pre := LastDeck{Kind: LastDeckPrebuilt, ID: "izzet-aggro"}
	if err := s.SetLastDeck(ctx, id, pre); err != nil {
		t.Fatalf("SetLastDeck(prebuilt): %v", err)
	}
	if got, err := s.LastDeck(ctx, id); err != nil || got != pre {
		t.Errorf("LastDeck = %+v, %v; want %+v (the last write wins)", got, err, pre)
	}
}

func TestSetLastDeckRefusesWhatItWillNotStore(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t, nil)
	id := signedInUser(t, s)
	for name, bad := range map[string]LastDeck{
		"zero":           {},
		"unknown kind":   {Kind: "url", ID: "x"},
		"empty id":       {Kind: LastDeckPrebuilt},
		"long id":        {Kind: LastDeckPrebuilt, ID: strings.Repeat("x", MaxLastDeckIDLen+1)},
		"library no uid": {Kind: LastDeckLibrary, ID: "not-a-uuid"},
	} {
		if err := s.SetLastDeck(ctx, id, bad); !errors.Is(err, ErrInvalidLastDeck) {
			t.Errorf("%s: SetLastDeck = %v, want ErrInvalidLastDeck", name, err)
		}
	}
	var raw *string
	if err := d.QueryRow(`SELECT last_deck FROM users WHERE id = ?`, id.String()).Scan(&raw); err != nil || raw != nil {
		t.Errorf("a refused write left last_deck = %v (%v)", raw, err)
	}
}

func TestLastDeckForAnUnknownUser(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	if _, err := s.LastDeck(ctx, uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("LastDeck = %v, want ErrNotFound", err)
	}
	if err := s.SetLastDeck(ctx, uuid.New(), LastDeck{Kind: LastDeckPrebuilt, ID: "x"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("SetLastDeck = %v, want ErrNotFound", err)
	}
}

// A value a newer binary wrote (an unknown kind) reads as none, not as
// an error: it costs a preselection and nothing else.
func TestLastDeckReadsAnUnknownStoredKindAsNone(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t, nil)
	id := signedInUser(t, s)
	for _, raw := range []string{`{"kind":"cloud","id":"x"}`, `not json`} {
		if _, err := d.Exec(`UPDATE users SET last_deck = ? WHERE id = ?`, raw, id.String()); err != nil {
			t.Fatal(err)
		}
		if got, err := s.LastDeck(ctx, id); err != nil || !got.IsZero() {
			t.Errorf("stored %s: LastDeck = %+v, %v; want none", raw, got, err)
		}
	}
}

func TestLastDeckIsPerUser(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	a := signedInUser(t, s)
	b, err := s.UpsertFromDiscord(ctx, alice2, "", "identify")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetLastDeck(ctx, a, LastDeck{Kind: LastDeckPrebuilt, ID: "x"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.LastDeck(ctx, b.ID); !got.IsZero() {
		t.Errorf("another user's last deck = %+v", got)
	}
}

func TestNoStoreLastDeck(t *testing.T) {
	var s Store = NoStore{}
	if _, err := s.LastDeck(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("LastDeck = %v", err)
	}
	if err := s.SetLastDeck(context.Background(), uuid.New(), LastDeck{Kind: LastDeckPrebuilt, ID: "x"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("SetLastDeck = %v", err)
	}
}
