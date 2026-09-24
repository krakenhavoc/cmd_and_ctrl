package game

import (
	"testing"

	"github.com/google/uuid"
)

// dropped_card_set_pick_test.go — #1225, the third shape of the
// departure table's dropDefault action.
//
// #1006 taught the drop to run a dropped OPTION PICK's continuation,
// and #1027 taught it to settle a dropped RUN LEG. The shape between
// them went unserved: a choose_cards asked from inside a resolution
// that is paused waiting for it, carrying nothing but a
// chooseCardsFrame. effects.SacrificeChoice is exactly that — "Torment
// of Hailfire repeats X times, and the next repetition must not be
// asked until this one has finished" — so a withdrawn sacrifice pick
// left the card stopped at the victim whose board emptied, which is
// the sentence #1006 wrote about the option-pick half of the very same
// card.
//
// The drop now runs the frame with NOTHING PICKED. That is not a new
// outcome invented here: QueueChooseCardsForEffect has no
// empty-candidate short-circuit, so every one of these callers already
// wrote the "there was nothing to ask" path itself and it runs the
// same closure with the same empty answer. What is new is that the
// DROP path reaches it.

// midCardCardPick queues the shape effects.SacrificeChoice queues: a
// mandatory one-card pick over the chooser's own battlefield whose
// continuation is the rest of the card. It records what it was run
// with.
func midCardCardPick(t *testing.T, g *Game, chooser, source uuid.UUID, cards []uuid.UUID) (uuid.UUID, *cardSetPickCall) {
	t.Helper()
	call := &cardSetPickCall{}
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:    chooser,
			FromPlayer: chooser,
			Source:     source,
			Question:   "Sacrifice a nonland permanent",
			Cards:      cards,
			Min:        1,
			Max:        1,
			Zone:       ZoneBattlefield,
			Then: func(_ *Game, picked []uuid.UUID) error {
				call.ran++
				call.got = picked
				return nil
			},
		})
	})
	if id == uuid.Nil {
		t.Fatal("setup: the mid-card pick was not queued")
	}
	return id, call
}

// TestAWithdrawnMidCardCardPickStillRunsTheRestOfTheCard is the issue's
// own shape: the chooser is still at the table, and the prune withdraws
// the question because the board it was about emptied under it.
//
// The floor of one is what makes it reachable. A zero-floor pick keeps
// an answer when its list shrinks, so the prune leaves it alone; a
// mandatory one has no answer left and has to go.
func TestAWithdrawnMidCardCardPickStillRunsTheRestOfTheCard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Torment of Hailfire")
	bear := permanentOf(g, victim.ID, "Their Bear")

	id, call := midCardCardPick(t, g, victim.ID, source, []uuid.UUID{bear})

	// Their last nonland permanent leaves while the question is open.
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	if findChoice(g, id) != nil {
		t.Fatal("the prompt survives a board with none of its candidates on it")
	}
	if call.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — the rest of the card was dropped "+
			"with the question (#1225)", call.ran)
	}
	if len(call.got) != 0 {
		t.Errorf("the continuation was handed %v; a withdrawal sacrifices nothing "+
			"on the chooser's behalf", call.got)
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, victim.ID) == nil {
		t.Error("the withdrawal is not in the log; a stall dump has to say where the prompt went")
	}
	assertTableIsFree(t, g)
}

// TestADepartedChoosersMidCardCardPickStillRunsTheRestOfTheCard is the
// other door into the same action: the chooser LEAVES with the question
// in front of them.
//
// The prompt is about their own board, so CR 800.4g reassigns nothing
// (FromPlayer is the chooser) and the departure sweep drops it — and
// the card whose text the frame belongs to is a survivor's, so gate 2
// passes and the rest of that card runs.
func TestADepartedChoosersMidCardCardPickStillRunsTheRestOfTheCard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Torment of Hailfire")
	bear := permanentOf(g, leaver.ID, "Their Bear")

	id, call := midCardCardPick(t, g, leaver.ID, source, []uuid.UUID{bear})

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if c := findChoice(g, id); c != nil {
		t.Fatalf("a pick over the departed seat's own board was reassigned to %s", c.Chooser)
	}
	if call.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", call.ran)
	}
	if len(call.got) != 0 {
		t.Errorf("the continuation was handed %v, want nothing", call.got)
	}
}

// TestADepartedChoosersOwnCardsMidCardPickRunsNothing is gate 2, the
// one the dropDefault action passes and nothing else does
// (choiceObjectSurvivesLocked). When the card the continuation belongs
// to left the game with the chooser, the resolution is over and there
// is no rest of the card to run.
func TestADepartedChoosersOwnCardsMidCardPickRunsNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	source := departureTestSource(g, leaver.ID, "Their Own Sorcery")
	bear := permanentOf(g, leaver.ID, "Their Bear")

	_, call := midCardCardPick(t, g, leaver.ID, source, []uuid.UUID{bear})

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if call.ran != 0 {
		t.Errorf("the continuation of a card that left the game with its controller ran %d times",
			call.ran)
	}
}

// TestAPartlyEmptiedMidCardPickIsTrimmedAndNotRunYet is the boundary,
// pinned so the fix is not read as "a choose_cards now ends whenever
// its board moves". The drop action runs on a WITHDRAWAL; a pick that
// still has a candidate is trimmed and left standing, and its
// continuation runs when the chooser answers and not before.
func TestAPartlyEmptiedMidCardPickIsTrimmedAndNotRunYet(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Torment of Hailfire")
	first := permanentOf(g, victim.ID, "First Bear")
	second := permanentOf(g, victim.ID, "Second Bear")

	id, call := midCardCardPick(t, g, victim.ID, source, []uuid.UUID{first, second})

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(first); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	c := findChoice(g, id)
	if c == nil {
		t.Fatal("one candidate is still on the battlefield; the prompt must not be withdrawn")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != second {
		t.Errorf("candidates after the prune = %v, want only the permanent still there", c.ChooseCards)
	}
	if call.ran != 0 {
		t.Fatalf("the continuation ran %d times with the question still open, want 0", call.ran)
	}

	if err := g.ResolveChooseCards(id, victim.ID, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if call.ran != 1 || len(call.got) != 1 || call.got[0] != second {
		t.Errorf("an ANSWERED pick runs the same continuation with the picks: ran=%d got=%v",
			call.ran, call.got)
	}
}

// TestAZeroFloorMidCardPickAlsoRunsOnItsWithdrawal — the floor decides
// whether the prune can ever empty the list under a real card, not what
// the drop does once it has. A zero-floor pick withdrawn with nothing
// left runs its continuation with nothing picked, like any other.
func TestAZeroFloorMidCardPickAlsoRunsOnItsWithdrawal(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Armored Skyhunter")
	bear := permanentOf(g, me.ID, "Bear")

	call := &cardSetPickCall{}
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  me.ID,
			Source:   source,
			Question: "You may choose a creature",
			Cards:    []uuid.UUID{bear},
			Min:      0,
			Max:      1,
			Zone:     ZoneBattlefield,
			Then: func(_ *Game, picked []uuid.UUID) error {
				call.ran++
				call.got = picked
				return nil
			},
		})
	})

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	if findChoice(g, id) != nil {
		t.Fatal("a pick with an empty candidate list is withdrawn whatever its floor")
	}
	if call.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", call.ran)
	}
	if len(call.got) != 0 {
		t.Errorf("the continuation was handed %v, want nothing", call.got)
	}
}

// TestAWithdrawnPickWithNoContinuationStillRunsNothing — the
// Thoughtseize-shaped pick the departure table's comment protects. A
// frame with no `then` has no rest of the card, and the drop must not
// invent one.
func TestAWithdrawnPickWithNoContinuationStillRunsNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Thoughtseize")
	bear := permanentOf(g, me.ID, "Bear")

	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  me.ID,
			Source:   source,
			Question: "Choose a card",
			Cards:    []uuid.UUID{bear},
			Min:      1,
			Max:      1,
			Zone:     ZoneBattlefield,
		})
	})
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})
	if findChoice(g, id) != nil {
		t.Error("the prompt survives with none of its candidates in the zone")
	}
	assertTableIsFree(t, g)
}
