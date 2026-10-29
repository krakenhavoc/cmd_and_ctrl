package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// may_cast_keyed_test.go — #4361: a may_cast's answer is a registered key,
// so a table waiting on one is a restore point, the restored answer reaches
// the right branch with the prompt's data, and a binary without the key
// refuses the file.

var (
	keyedMayCastSeen []MayCastAnswer
	keyedMayCast     = RegisterMayCastThen(testEffectKeyPrefix+"keyed-may-cast", func(_ *Game, r MayCastAnswer) error {
		keyedMayCastSeen = append(keyedMayCastSeen, r)
		return nil
	})
)

func queueKeyedMayCast(t *testing.T, g *Game, chooser, source, card, carry uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueueMayCastKeyedForEffect(MayCastPrompt{
			Chooser:  chooser,
			Source:   source,
			Card:     card,
			Question: "cast it without paying its mana cost?",
			Keyword:  MayCastKeywordPlayOrHand,
		}, keyedMayCast, []uuid.UUID{carry}); err != nil {
			t.Fatal(err)
		}
	})
}

func mayCastChoiceFor(g *Game, chooser uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestKeyedMayCastSurvivesARestoreAndAnswersBothWays(t *testing.T) {
	for _, accept := range []bool{true, false} {
		keyedMayCastSeen = nil
		g := newActiveGame(t)
		me := g.Seats[0]
		source, card, carried := uuid.New(), uuid.New(), uuid.New()
		queueKeyedMayCast(t, g, me.ID, source, card, carried)

		restored := restoreStrictForTest(t, g)
		c := mayCastChoiceFor(restored, me.ID)
		if c == nil {
			t.Fatal("the restored table lost the offer")
		}
		if c.MayCastCard != card || c.MayCastKeyword != MayCastKeywordPlayOrHand {
			t.Errorf("restored offer over %v under %q, want %v under %q", c.MayCastCard, c.MayCastKeyword, card, MayCastKeywordPlayOrHand)
		}
		if err := restored.ResolveMayCast(c.ID, me.ID, accept); err != nil {
			t.Fatalf("ResolveMayCast on the restored game: %v", err)
		}
		if len(keyedMayCastSeen) != 1 {
			t.Fatalf("accept=%v: the answer ran %d times, want 1", accept, len(keyedMayCastSeen))
		}
		r := keyedMayCastSeen[0]
		if r.Accepted != accept || r.Chooser != me.ID || r.Source != source || r.Card != card ||
			len(r.Carry) != 1 || r.Carry[0] != carried {
			t.Errorf("accept=%v: answer got %+v", accept, r)
		}
		if mayCastChoiceFor(restored, me.ID) != nil {
			t.Error("the answered offer is still open")
		}
	}
}

func TestKeyedMayCastRefusesAnUnknownKeyAtRestore(t *testing.T) {
	g := newActiveGame(t)
	queueKeyedMayCast(t, g, g.Seats[0].ID, uuid.New(), uuid.New(), uuid.New())
	snap := g.CaptureSnapshot()
	for i := range snap.PendingChoices {
		snap.PendingChoices[i].PickThen = "test/never-registered"
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown may-cast answer: err = %v, want ErrUnknownEffectKey", err)
	}
}

// A chooser who has left the game is not asked: the decline runs at once,
// with the prompt's data.
func TestKeyedMayCastForAChooserWhoHasLeftRunsTheDecline(t *testing.T) {
	keyedMayCastSeen = nil
	g := newActiveGame(t)
	gone := g.Seats[1]
	g.WithWriteLock(func() { gone.Eliminated = true })
	card := uuid.New()
	queueKeyedMayCast(t, g, gone.ID, uuid.New(), card, uuid.New())
	if mayCastChoiceFor(g, gone.ID) != nil {
		t.Fatal("a player who has left was asked")
	}
	if len(keyedMayCastSeen) != 1 || keyedMayCastSeen[0].Accepted || keyedMayCastSeen[0].Card != card {
		t.Fatalf("the decline did not run once with the card: %+v", keyedMayCastSeen)
	}
}

func TestKeyedMayCastWithNoAnswerIsRefused(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		err := g.QueueMayCastKeyedForEffect(MayCastPrompt{Chooser: g.Seats[0].ID, Card: uuid.New()}, MayCastThen{}, nil)
		if !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("a keyed may-cast with no answer: err = %v, want ErrInvalidParam", err)
		}
	})
	if len(g.PendingChoices) != 0 {
		t.Errorf("a refused offer was queued: %+v", g.PendingChoices)
	}
}
