package decklibrary

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// Tests for ADR 0110's additions: decks.source_url, the 200-deck cap,
// delete (which nulls seats.deck_id) and rename.

func mustSeat(t *testing.T, d *db.DB, game string, seat int, deckID *uuid.UUID) {
	t.Helper()
	if _, err := d.Exec(`INSERT OR IGNORE INTO games (id, name, state, created_at) VALUES (?, 'G', 'lobby', 1)`, game); err != nil {
		t.Fatal(err)
	}
	var deck any
	if deckID != nil {
		deck = deckID.String()
	}
	if _, err := d.Exec(`INSERT INTO seats (game_id, seat, player_id, deck_id, deck_name) VALUES (?, ?, ?, ?, 'Label')`,
		game, seat, fmt.Sprintf("p%d", seat), deck); err != nil {
		t.Fatalf("insert seat: %v", err)
	}
}

func seatDeck(t *testing.T, d *db.DB, game string, seat int) (deckID *string, deckName *string) {
	t.Helper()
	if err := d.QueryRow(`SELECT deck_id, deck_name FROM seats WHERE game_id = ? AND seat = ?`, game, seat).Scan(&deckID, &deckName); err != nil {
		t.Fatalf("read seat: %v", err)
	}
	return
}

func TestUpsertFromLinkSavesTheLinkBesideTheList(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")

	deck, err := s.UpsertFromLink(ctx, owner, "From Moxfield", "moxfield", `{"boards":{}}`, "https://moxfield.com/decks/abc", []string{"Atraxa"}, 100)
	if err != nil {
		t.Fatalf("UpsertFromLink: %v", err)
	}
	if deck.SourceURL != "https://moxfield.com/decks/abc" || deck.SourceFormat != "moxfield" || deck.SourceText != `{"boards":{}}` {
		t.Errorf("deck = %+v", deck)
	}
	got, err := s.Get(ctx, deck.ID)
	if err != nil || got.SourceURL != deck.SourceURL {
		t.Errorf("Get = %+v, %v", got, err)
	}
	list, _ := s.List(ctx, owner)
	if len(list) != 1 || list[0].SourceURL != deck.SourceURL {
		t.Errorf("List = %+v", list)
	}
}

func TestAPastedDeckHasNoSourceURLAndResavingFromTextClearsIt(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")

	pasted, err := s.Upsert(ctx, owner, "Deck", "text", "1 Atraxa", nil, 100)
	if err != nil || pasted.SourceURL != "" {
		t.Fatalf("Upsert = %+v, %v", pasted, err)
	}
	var raw *string
	if err := d.QueryRow(`SELECT source_url FROM decks WHERE id = ?`, pasted.ID.String()).Scan(&raw); err != nil || raw != nil {
		t.Errorf("stored source_url = %v (%v), want NULL", raw, err)
	}

	linked, err := s.UpsertFromLink(ctx, owner, "Deck", "text", "1 Atraxa\n1 Sol Ring", "https://example.test/d", nil, 100)
	if err != nil || linked.ID != pasted.ID || linked.SourceURL != "https://example.test/d" {
		t.Fatalf("link over paste = %+v, %v (same name updates in place)", linked, err)
	}
	again, err := s.Upsert(ctx, owner, "Deck", "text", "1 Atraxa", nil, 100)
	if err != nil || again.ID != pasted.ID || again.SourceURL != "" {
		t.Errorf("paste over link = %+v, %v; want the link cleared", again, err)
	}
}

func TestTheDeckCapIs200AndOnlyRefusesInserts(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	owner, other := mustUser(t, d, "Alice"), mustUser(t, d, "Bob")

	for i := 0; i < MaxDecks; i++ {
		if _, err := s.Upsert(ctx, owner, fmt.Sprintf("Deck %d", i), "text", "1 Atraxa", nil, 100); err != nil {
			t.Fatalf("deck %d: %v", i, err)
		}
	}
	if n, err := s.Count(ctx, owner); err != nil || n != MaxDecks {
		t.Fatalf("Count = %d, %v; want %d", n, err, MaxDecks)
	}
	if _, err := s.Upsert(ctx, owner, "One too many", "text", "x", nil, 100); !errors.Is(err, ErrLibraryFull) {
		t.Fatalf("the 201st deck = %v, want ErrLibraryFull", err)
	}
	if _, err := s.UpsertFromLink(ctx, owner, "Link too many", "text", "x", "https://example.test", nil, 100); !errors.Is(err, ErrLibraryFull) {
		t.Fatalf("the 201st deck, from a link = %v, want ErrLibraryFull", err)
	}
	if n, _ := s.Count(ctx, owner); n != MaxDecks {
		t.Errorf("a refused save changed the count to %d", n)
	}
	// Updating a deck already in a full library is fine.
	if _, err := s.Upsert(ctx, owner, "Deck 7", "text", "1 Sol Ring", nil, 99); err != nil {
		t.Errorf("updating in a full library: %v", err)
	}
	// The cap is per person.
	if _, err := s.Upsert(ctx, other, "Mine", "text", "x", nil, 100); err != nil {
		t.Errorf("another person's first deck: %v", err)
	}
	// Deleting one frees a slot.
	list, _ := s.List(ctx, owner)
	if err := s.Delete(ctx, owner, list[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Upsert(ctx, owner, "Fits now", "text", "x", nil, 100); err != nil {
		t.Errorf("after a delete: %v", err)
	}
}

func TestDeleteRemovesTheDeckAndNullsEverySeatThatUsedIt(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")
	doomed, err := s.Upsert(ctx, owner, "Doomed", "text", "x", nil, 100)
	if err != nil {
		t.Fatal(err)
	}
	keep, err := s.Upsert(ctx, owner, "Keep", "text", "x", nil, 100)
	if err != nil {
		t.Fatal(err)
	}
	mustSeat(t, d, "g1", 0, &doomed.ID)
	mustSeat(t, d, "g1", 1, &keep.ID)
	mustSeat(t, d, "g2", 0, &doomed.ID)
	mustSeat(t, d, "g2", 1, nil)

	if err := s.Delete(ctx, owner, doomed.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get(ctx, doomed.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after delete = %v", err)
	}
	for _, c := range []struct {
		game string
		seat int
	}{{"g1", 0}, {"g2", 0}} {
		id, name := seatDeck(t, d, c.game, c.seat)
		if id != nil || name == nil || *name != "Label" {
			t.Errorf("seat %s/%d: deck_id = %v, deck_name = %v; want NULL and the label kept", c.game, c.seat, id, name)
		}
	}
	if id, _ := seatDeck(t, d, "g1", 1); id == nil || *id != keep.ID.String() {
		t.Errorf("a seat on another deck lost it: %v", id)
	}
	if n, _ := s.Count(ctx, owner); n != 1 {
		t.Errorf("Count = %d, want 1", n)
	}
}

func TestDeleteOfSomeoneElsesOrAMissingDeckIsNotFound(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	owner, thief := mustUser(t, d, "Alice"), mustUser(t, d, "Mallory")
	deck, err := s.Upsert(ctx, owner, "Mine", "text", "x", nil, 100)
	if err != nil {
		t.Fatal(err)
	}
	mustSeat(t, d, "g1", 0, &deck.ID)

	if err := s.Delete(ctx, thief, deck.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete of another person's deck = %v, want ErrNotFound", err)
	}
	if err := s.Delete(ctx, owner, uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete of a missing deck = %v, want ErrNotFound", err)
	}
	if _, err := s.Get(ctx, deck.ID); err != nil {
		t.Errorf("the deck was lost: %v", err)
	}
	if id, _ := seatDeck(t, d, "g1", 0); id == nil {
		t.Error("a refused delete nulled the seat")
	}
}

// The schema's own ON DELETE SET NULL covers a delete that does not go
// through the store.
func TestTheSchemaNullsSeatsOnARawDelete(t *testing.T) {
	s, d := openStore(t)
	owner := mustUser(t, d, "Alice")
	deck, err := s.Upsert(context.Background(), owner, "Deck", "text", "x", nil, 100)
	if err != nil {
		t.Fatal(err)
	}
	mustSeat(t, d, "g1", 0, &deck.ID)
	if _, err := d.Exec(`DELETE FROM decks WHERE id = ?`, deck.ID.String()); err != nil {
		t.Fatalf("raw delete: %v", err)
	}
	if id, _ := seatDeck(t, d, "g1", 0); id != nil {
		t.Errorf("deck_id = %v, want NULL", *id)
	}
}

func TestRename(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	owner, other := mustUser(t, d, "Alice"), mustUser(t, d, "Bob")
	a, err := s.Upsert(ctx, owner, "A", "text", "x", nil, 100)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Upsert(ctx, owner, "B", "text", "x", nil, 100); err != nil {
		t.Fatal(err)
	}
	theirs, err := s.Upsert(ctx, other, "A", "text", "x", nil, 100)
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.Rename(ctx, owner, a.ID, "A2")
	if err != nil || got.Name != "A2" || !got.UpdatedAt.Equal(a.UpdatedAt) {
		t.Fatalf("Rename = %+v, %v; want the new name and an unchanged updated_at", got, err)
	}
	if read, _ := s.Get(ctx, a.ID); read.Name != "A2" {
		t.Errorf("stored name = %q", read.Name)
	}
	if _, err := s.Rename(ctx, owner, a.ID, "B"); !errors.Is(err, ErrNameTaken) {
		t.Errorf("rename onto another deck's name = %v, want ErrNameTaken", err)
	}
	if _, err := s.Rename(ctx, owner, a.ID, "A2"); err != nil {
		t.Errorf("rename to its own name: %v", err)
	}
	// Another person's deck of the same name does not collide.
	if _, err := s.Rename(ctx, owner, a.ID, "A"); err != nil {
		t.Errorf("rename to a name only someone else uses: %v", err)
	}
	if _, err := s.Rename(ctx, owner, theirs.ID, "Stolen"); !errors.Is(err, ErrNotFound) {
		t.Errorf("rename of another person's deck = %v, want ErrNotFound", err)
	}
	if _, err := s.Rename(ctx, owner, uuid.New(), "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("rename of a missing deck = %v, want ErrNotFound", err)
	}
	if _, err := s.Rename(ctx, owner, a.ID, ""); err == nil {
		t.Error("rename to an empty name succeeded")
	}
	if read, _ := s.Get(ctx, theirs.ID); read.Name != "A" {
		t.Errorf("another person's deck was renamed to %q", read.Name)
	}
}

func TestNoStoreRememberMeMethods(t *testing.T) {
	ctx := context.Background()
	var s Store = NoStore{}
	if n, err := s.Count(ctx, uuid.New()); n != 0 || err != nil {
		t.Errorf("Count = %d, %v", n, err)
	}
	if err := s.Delete(ctx, uuid.New(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete = %v", err)
	}
	if _, err := s.Rename(ctx, uuid.New(), uuid.New(), "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Rename = %v", err)
	}
	if _, err := s.UpsertFromLink(ctx, uuid.New(), "x", "text", "x", "u", nil, 1); err == nil {
		t.Error("UpsertFromLink = nil, want an error")
	}
}
