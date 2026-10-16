package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// choose_cards_keyed_test.go — #2040: a choose_cards pick whose
// continuation is a registered key survives a restore, runs with nobody's
// choice when it is dropped, and refuses a key this binary lacks.

var (
	keyedCardsSeen []CardsChosen
	keyedCards     = RegisterChooseCardsThen(testEffectKeyPrefix+"keyed-choose-cards", func(_ *Game, r CardsChosen) error {
		keyedCardsSeen = append(keyedCardsSeen, r)
		return nil
	})
)

func queueKeyedChooseCards(t *testing.T, g *Game, chooser uuid.UUID, cards []uuid.UUID, carry uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  chooser,
			Source:   uuid.New(),
			Question: "pick one",
			Cards:    cards,
			Min:      1,
			Max:      1,
			ThenKey:  keyedCards,
			Carry:    []uuid.UUID{carry},
		})
	})
	if id == uuid.Nil {
		t.Fatal("nothing was queued")
	}
	return id
}

func TestKeyedChooseCardsSurvivesARestore(t *testing.T) {
	keyedCardsSeen = nil
	g := newActiveGame(t)
	me := g.Seats[0]
	cards := []uuid.UUID{uuid.New(), uuid.New()}
	carried := uuid.New()
	id := queueKeyedChooseCards(t, g, me.ID, cards, carried)
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a keyed choose_cards pick blocks the restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if err := restored.ResolveChooseCards(id, me.ID, []uuid.UUID{cards[1]}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored game: %v", err)
	}
	if len(keyedCardsSeen) != 1 {
		t.Fatalf("continuation ran %d times, want 1", len(keyedCardsSeen))
	}
	r := keyedCardsSeen[0]
	if !r.Answered || len(r.Picked) != 1 || r.Picked[0] != cards[1] || len(r.Candidates) != 2 ||
		r.Min != 1 || len(r.Carry) != 1 || r.Carry[0] != carried {
		t.Errorf("continuation got %+v", r)
	}
}

func TestKeyedChooseCardsRunsWithNoChoiceWhenDropped(t *testing.T) {
	keyedCardsSeen = nil
	g := newActiveGame(t)
	me := g.Seats[1]
	queueKeyedChooseCards(t, g, me.ID, []uuid.UUID{uuid.New()}, uuid.New())
	g.WithWriteLock(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == me.ID {
				g.defaultDroppedChoiceLocked(c)
			}
		}
	})
	if len(keyedCardsSeen) != 1 || keyedCardsSeen[0].Answered || keyedCardsSeen[0].Picked != nil {
		t.Fatalf("dropped: continuation got %+v, want one unanswered run", keyedCardsSeen)
	}
}

func TestKeyedChooseCardsRefusesAnUnknownKeyAtRestore(t *testing.T) {
	g := newActiveGame(t)
	queueKeyedChooseCards(t, g, g.Seats[0].ID, []uuid.UUID{uuid.New()}, uuid.New())
	snap := g.CaptureSnapshot()
	for i := range snap.PendingChoices {
		snap.PendingChoices[i].PickThen = "test/never-registered"
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown choose-cards continuation: err = %v, want ErrUnknownEffectKey", err)
	}
}
