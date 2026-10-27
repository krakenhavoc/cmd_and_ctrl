package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exile_return_then_test.go — #1327: the exile return with a
// continuation. "If it entered under your control" (Phelia, Exuberant
// Shepherd) can only be asked once the entry is complete, which is one
// action later when the entry stopped for a question.

// TestExileReturnThenWaitsForAPausedEntry: the returning card's own
// entry asks a question, so nothing has entered when the call returns.
// The continuation must not run then — it runs from the answer, told
// the NEW object's ID (CR 400.7), with the permanent already there.
func TestExileReturnThenWaitsForAPausedEntry(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := ebLand(me.ID, "Fountain")
	g.Exile.PushTop(card)

	resetProbes()
	testEnteredSeen = nil
	g.mu.Lock()
	ebPayLifeToUntap(g, card.InstanceID, me.ID, 2)
	err := g.ReturnFromExileToBattlefieldKeyedForEffect(card.InstanceID, uuid.Nil, false, KeyedTail{Ref: testEnteredTail})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("return: %v", err)
	}
	if len(testProbeHeard) != 0 {
		t.Fatal("the continuation ran while the entry's question was open")
	}
	c := ebPayLifePrompt(g)
	if c == nil {
		t.Fatal("the returning card's question was not asked")
	}
	if err := g.ResolveEntryPayLife(c.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if len(testProbeHeard) != 1 {
		t.Fatalf("continuation ran %d times, want 1", len(testProbeHeard))
	}
	got, sawOnBattlefield := probeEntered()[0], testEnteredSeen[0]
	if got == uuid.Nil || got == card.InstanceID {
		t.Errorf("continuation told %s; want the new object's ID (CR 400.7)", got)
	}
	if !sawOnBattlefield {
		t.Error("the continuation ran before the permanent was on the battlefield")
	}
}

// TestExileReturnThenIsToldWhenNothingEnters covers the terminal
// outcomes with nothing on the battlefield: a cancelled entry, and a
// card that is not in exile at all. The continuation is still told —
// with uuid.Nil — because a caller sequenced behind it would otherwise
// wait forever.
func TestExileReturnThenIsToldWhenNothingEnters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := ebLand(me.ID, "Fountain")
	g.Exile.PushTop(card)

	resetProbes()
	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == card.InstanceID && ev.NewZone == ZoneBattlefield
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		Label: "test: cancel the return",
	})
	if err := g.ReturnFromExileToBattlefieldKeyedForEffect(card.InstanceID, uuid.Nil, false, probeTail()); err != nil {
		t.Errorf("cancelled return: %v", err)
	}
	err := g.ReturnFromExileToBattlefieldKeyedForEffect(uuid.New(), uuid.Nil, false, probeTail())
	g.mu.Unlock()
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("missing card: err = %v, want ErrCardNotFound", err)
	}
	got := probeEntered()
	if calls := len(got); calls != 2 || got[0] != uuid.Nil || got[1] != uuid.Nil {
		t.Errorf("continuation calls %d with %v; want 2 with uuid.Nil", len(got), got)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("a cancelled return leaves the card in exile")
	}
}
