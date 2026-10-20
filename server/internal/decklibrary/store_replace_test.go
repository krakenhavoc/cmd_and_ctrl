package decklibrary

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// ReplaceContents (#2036) is the write behind updating a saved deck.

func TestReplaceContentsKeepsIdentityAndReplacesTheRest(t *testing.T) {
	s, d := openStore(t)
	ctx := context.Background()
	owner := mustUser(t, d, "Alice")
	before, err := s.UpsertFromLink(ctx, owner, "Atraxa", "text", "v1", "https://archidekt.com/decks/1", []string{"Atraxa"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetPowerLevel(ctx, before.ID, 4); err != nil {
		t.Fatal(err)
	}
	before, _ = s.Get(ctx, before.ID)

	got, err := s.ReplaceContents(ctx, before, "text", "v2", "", []string{"Atraxa", "Tymna"}, 99)
	if err != nil {
		t.Fatalf("ReplaceContents: %v", err)
	}
	if got.ID != before.ID || got.Name != "Atraxa" || got.OwnerID != owner || !got.CreatedAt.Equal(before.CreatedAt) {
		t.Errorf("identity changed: %+v", got)
	}
	if got.SourceText != "v2" || got.SourceURL != "" || got.CardCount != 99 || len(got.Commanders) != 2 {
		t.Errorf("contents not replaced: %+v", got)
	}
	if got.PowerLevel != 0 {
		t.Errorf("rating = %v, want cleared with the list", got.PowerLevel)
	}
	if !got.UpdatedAt.After(before.UpdatedAt) {
		t.Errorf("updated_at = %v, want after %v", got.UpdatedAt, before.UpdatedAt)
	}
}

func TestReplaceContentsRefusesAStaleEditorAndAStranger(t *testing.T) {
	s, d := openStore(t)
	ctx := context.Background()
	owner := mustUser(t, d, "Alice")
	before, err := s.Upsert(ctx, owner, "Deck", "text", "1 Plains", []string{}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Upsert(ctx, owner, "Deck", "text", "2 Plains", []string{}, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReplaceContents(ctx, before, "text", "3 Plains", "", []string{}, 3); !errors.Is(err, ErrSourceChanged) {
		t.Fatalf("stale source: %v", err)
	}
	before.OwnerID = uuid.New()
	if _, err := s.ReplaceContents(ctx, before, "text", "3 Plains", "", []string{}, 3); !errors.Is(err, ErrNotFound) {
		t.Fatalf("wrong owner: %v", err)
	}
	current, _ := s.Get(ctx, before.ID)
	if current.SourceText != "2 Plains" {
		t.Errorf("the newer save was overwritten: %q", current.SourceText)
	}
}

func TestReplaceContentsKeepsTheRatingForAnIdenticalList(t *testing.T) {
	s, d := openStore(t)
	ctx := context.Background()
	owner := mustUser(t, d, "Alice")
	deck, _ := s.Upsert(ctx, owner, "Deck", "text", "v1", []string{}, 1)
	if _, err := s.SetPowerLevel(ctx, deck.ID, 3); err != nil {
		t.Fatal(err)
	}
	deck, _ = s.Get(ctx, deck.ID)
	got, err := s.ReplaceContents(ctx, deck, "text", "v1", "https://archidekt.com/decks/9", []string{}, 1)
	if err != nil || got.PowerLevel != 3 || got.SourceURL == "" {
		t.Errorf("same list, new link: %+v %v", got, err)
	}
}

func TestNoStoreReplaceContentsIsNotFound(t *testing.T) {
	if _, err := (NoStore{}).ReplaceContents(context.Background(), Deck{}, "text", "", "", nil, 0); !errors.Is(err, ErrNotFound) {
		t.Errorf("NoStore.ReplaceContents = %v", err)
	}
}
