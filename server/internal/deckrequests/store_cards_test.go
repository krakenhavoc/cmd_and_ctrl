package deckrequests

import (
	"context"
	"errors"
	"testing"
)

func TestRecordedCardsAreKeptPerDeckAndNeverListedTwice(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()

	if got, err := s.CardsOf(ctx, "moxfield:a"); err != nil || len(got) != 0 {
		t.Fatalf("CardsOf an unrequested deck = %v, %v", got, err)
	}
	if err := s.RecordCards(ctx, "moxfield:a", 7, []Card{{Key: "o-2", Name: "Zed"}, {Key: "o-1", Name: "Alpha"}}); err != nil {
		t.Fatal(err)
	}
	// A second issue listing a card the deck already has leaves the
	// first row alone; a card it did not have is added.
	if err := s.RecordCards(ctx, "moxfield:a", 9, []Card{{Key: "o-1", Name: "Alpha"}, {Key: "o-3", Name: "Mid"}}); err != nil {
		t.Fatal(err)
	}
	got, err := s.CardsOf(ctx, "moxfield:a")
	if err != nil {
		t.Fatal(err)
	}
	want := []Card{{"o-1", "Alpha", 7}, {"o-3", "Mid", 9}, {"o-2", "Zed", 7}}
	if len(got) != len(want) {
		t.Fatalf("cards = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("card %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if other, _ := s.CardsOf(ctx, "moxfield:b"); len(other) != 0 {
		t.Errorf("another deck's cards = %+v", other)
	}
	if err := s.RecordCards(ctx, "", 1, []Card{{Key: "x", Name: "X"}}); err == nil {
		t.Error("RecordCards accepted no deck key")
	}
	if err := s.RecordCards(ctx, "moxfield:a", 0, []Card{{Key: "x", Name: "X"}}); err == nil {
		t.Error("RecordCards accepted no issue number")
	}
}

func TestCopyCardsGivesTheNewKeyTheOldKeysHistory(t *testing.T) {
	s := openStore(t)
	ctx := context.Background()
	if err := s.RecordCards(ctx, "list:old", 7, []Card{{Key: "o-1", Name: "Alpha"}, {Key: "o-2", Name: "Beta"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordCards(ctx, "list:new", 9, []Card{{Key: "o-2", Name: "Beta"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.CopyCards(ctx, "list:old", "list:new"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.CardsOf(ctx, "list:new")
	if len(got) != 2 || got[0] != (Card{"o-1", "Alpha", 7}) || got[1] != (Card{"o-2", "Beta", 9}) {
		t.Errorf("after the copy = %+v: want Alpha from issue 7 and Beta still from issue 9", got)
	}
	if old, _ := s.CardsOf(ctx, "list:old"); len(old) != 2 {
		t.Errorf("the copy changed the source: %+v", old)
	}
	if err := s.CopyCards(ctx, "list:new", "list:new"); err != nil {
		t.Errorf("a copy onto itself failed: %v", err)
	}
	if err := (NoStore{}).CopyCards(ctx, "a", "b"); !errors.Is(err, ErrNoStore) {
		t.Errorf("NoStore.CopyCards = %v", err)
	}
}
