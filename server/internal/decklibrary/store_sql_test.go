package decklibrary

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func openStore(t *testing.T) (*SQLStore, *db.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return NewSQLStore(d), d
}

// mustUser inserts a users row directly (decks.owner_id is a real
// foreign key from migration 0005) and returns its id.
func mustUser(t *testing.T, d *db.DB, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, ?, ?)`,
		id.String(), name, 1, 1); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func clock(s *SQLStore, at time.Time) { s.now = func() time.Time { return at } }

func TestUpsertInsertsANewDeck(t *testing.T) {
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")
	t0 := time.UnixMilli(1_700_000_000_123).UTC()
	clock(s, t0)

	deck, err := s.Upsert(context.Background(), owner, "Atraxa Superfriends", "text", "1 Atraxa\n...", []string{"Atraxa, Praetors' Voice"}, 100)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if deck.ID == uuid.Nil {
		t.Fatal("new deck has no id")
	}
	if deck.OwnerID != owner || deck.Name != "Atraxa Superfriends" || deck.SourceFormat != "text" || deck.CardCount != 100 {
		t.Errorf("deck = %+v", deck)
	}
	if len(deck.Commanders) != 1 || deck.Commanders[0] != "Atraxa, Praetors' Voice" {
		t.Errorf("commanders = %v", deck.Commanders)
	}
	if !deck.CreatedAt.Equal(t0) || !deck.UpdatedAt.Equal(t0) {
		t.Errorf("timestamps = %v / %v, want %v", deck.CreatedAt, deck.UpdatedAt, t0)
	}
}

func TestUpsertSameOwnerAndNameUpdatesInPlace(t *testing.T) {
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")
	t0 := time.UnixMilli(1_700_000_000_000).UTC()
	clock(s, t0)

	first, err := s.Upsert(context.Background(), owner, "Atraxa Superfriends", "text", "v1", []string{"Atraxa, Praetors' Voice"}, 100)
	if err != nil {
		t.Fatalf("first Upsert: %v", err)
	}

	t1 := t0.Add(time.Hour)
	clock(s, t1)
	second, err := s.Upsert(context.Background(), owner, "Atraxa Superfriends", "moxfield", "v2", []string{"Atraxa, Praetors' Voice"}, 99)
	if err != nil {
		t.Fatalf("second Upsert: %v", err)
	}

	if second.ID != first.ID {
		t.Errorf("update rule: same owner+name got a new id (%s vs %s)", second.ID, first.ID)
	}
	if second.SourceFormat != "moxfield" || second.SourceText != "v2" || second.CardCount != 99 {
		t.Errorf("second = %+v, want the new source", second)
	}
	if !second.CreatedAt.Equal(t0) {
		t.Errorf("created_at moved on update: got %v, want %v", second.CreatedAt, t0)
	}
	if !second.UpdatedAt.Equal(t1) {
		t.Errorf("updated_at = %v, want %v", second.UpdatedAt, t1)
	}

	decks, err := s.List(context.Background(), owner)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(decks) != 1 {
		t.Fatalf("List after update: got %d decks, want 1 (update rule inserted a duplicate)", len(decks))
	}
}

func TestUpsertDifferentNameInsertsASecondDeck(t *testing.T) {
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")

	if _, err := s.Upsert(context.Background(), owner, "Deck A", "text", "a", []string{"A"}, 100); err != nil {
		t.Fatalf("Upsert A: %v", err)
	}
	if _, err := s.Upsert(context.Background(), owner, "Deck B", "text", "b", []string{"B"}, 100); err != nil {
		t.Fatalf("Upsert B: %v", err)
	}

	decks, err := s.List(context.Background(), owner)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(decks) != 2 {
		t.Fatalf("got %d decks, want 2", len(decks))
	}
}

func TestUpsertSameNameDifferentOwnerInsertsBoth(t *testing.T) {
	s, d := openStore(t)
	alice := mustUser(t, d, "Alice")
	bob := mustUser(t, d, "Bob")

	a, err := s.Upsert(context.Background(), alice, "Atraxa Superfriends", "text", "a", nil, 100)
	if err != nil {
		t.Fatalf("Upsert (alice): %v", err)
	}
	b, err := s.Upsert(context.Background(), bob, "Atraxa Superfriends", "text", "b", nil, 100)
	if err != nil {
		t.Fatalf("Upsert (bob): %v", err)
	}
	if a.ID == b.ID {
		t.Error("two different owners' decks with the same name collapsed into one row")
	}
}

func TestListOrdersNewestUpdatedFirst(t *testing.T) {
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")

	clock(s, time.UnixMilli(1000))
	oldest, err := s.Upsert(context.Background(), owner, "Old", "text", "x", nil, 100)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	clock(s, time.UnixMilli(2000))
	newest, err := s.Upsert(context.Background(), owner, "New", "text", "x", nil, 100)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	// Touch "Old" again so it becomes the most recently updated.
	clock(s, time.UnixMilli(3000))
	if _, err := s.Upsert(context.Background(), owner, "Old", "text", "x2", nil, 100); err != nil {
		t.Fatalf("Upsert (re-touch): %v", err)
	}

	decks, err := s.List(context.Background(), owner)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(decks) != 2 || decks[0].ID != oldest.ID || decks[1].ID != newest.ID {
		t.Fatalf("List order = %v, want [Old (re-touched), New]", decks)
	}
}

func TestListIsScopedToOwner(t *testing.T) {
	s, d := openStore(t)
	alice := mustUser(t, d, "Alice")
	bob := mustUser(t, d, "Bob")

	if _, err := s.Upsert(context.Background(), alice, "Alice's deck", "text", "a", nil, 100); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if _, err := s.Upsert(context.Background(), bob, "Bob's deck", "text", "b", nil, 100); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	decks, err := s.List(context.Background(), alice)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(decks) != 1 || decks[0].Name != "Alice's deck" {
		t.Fatalf("Alice's list = %v, want just her own deck", decks)
	}
}

func TestGetUnknownReturnsErrNotFound(t *testing.T) {
	s, _ := openStore(t)
	if _, err := s.Get(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get: got %v, want ErrNotFound", err)
	}
}

func TestGetRoundTripsCommanders(t *testing.T) {
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")
	saved, err := s.Upsert(context.Background(), owner, "Partners", "text", "x",
		[]string{"Tymna the Weaver", "Kraum, Ludevic's Opus"}, 100)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := s.Get(context.Background(), saved.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Commanders) != 2 || got.Commanders[0] != "Tymna the Weaver" || got.Commanders[1] != "Kraum, Ludevic's Opus" {
		t.Errorf("commanders round-trip = %v", got.Commanders)
	}
}

func TestUpsertRejectsNoOwnerOrName(t *testing.T) {
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")

	if _, err := s.Upsert(context.Background(), uuid.Nil, "Deck", "text", "x", nil, 100); err == nil {
		t.Error("Upsert with a zero owner: want an error")
	}
	if _, err := s.Upsert(context.Background(), owner, "", "text", "x", nil, 100); err == nil {
		t.Error("Upsert with an empty name: want an error")
	}
}

func TestNoStoreBehavesAsNoDecksForAnybody(t *testing.T) {
	var s NoStore
	if _, err := s.Upsert(context.Background(), uuid.New(), "Deck", "text", "x", nil, 100); err == nil {
		t.Error("NoStore.Upsert: want an error")
	}
	if _, err := s.Get(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("NoStore.Get: got %v, want ErrNotFound", err)
	}
	decks, err := s.List(context.Background(), uuid.New())
	if err != nil || len(decks) != 0 {
		t.Errorf("NoStore.List: got %v, %v, want (nil, nil)", decks, err)
	}
}
