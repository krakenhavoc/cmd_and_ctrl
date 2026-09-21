package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// entry_prune_test.go — #1069, the door #1045's prune did not watch.
//
// pruneCardSetChoicesLocked ran at three sites and all three are EXITS
// or a departure. A card that leaves a hand, a library or a graveyard
// FOR THE BATTLEFIELD takes none of them — routeDestinationLocked
// refuses a battlefield or stack destination outright — so a pick over
// the zone it left kept offering a card its own resolver would refuse,
// and an emptied one was the #544 wedge arriving through the one door
// nobody was standing at.
//
// The two shapes below are the two the issue names: a graveyard pick
// whose candidate is reanimated under the open prompt, and a hand pick
// whose candidate is put onto the battlefield by a Warp World-style
// effect. The third test is the sandbox landing, and the fourth is the
// reason the new call cannot double-run with the exit primitive's.

// permanentInHand puts a permanent card into p's hand — a card
// "put onto the battlefield" can name, which handCardForDiscard's
// sorcery cannot (CR 110.4).
func permanentInHand(t *testing.T, g *Game, p *Player, name string) uuid.UUID {
	t.Helper()
	c := NewCard(name, p.ID)
	c.TypeLine = "Creature — Bear"
	c.Power, c.Toughness = 2, 2
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

// permanentInGraveyard is the same card, one zone over.
func permanentInGraveyard(t *testing.T, g *Game, p *Player, name string) uuid.UUID {
	t.Helper()
	c := NewCard(name, p.ID)
	c.TypeLine = "Creature — Bear"
	c.Power, c.Toughness = 2, 2
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

// TestAGraveyardPickIsPrunedWhenItsCandidateIsReanimated is the issue's
// first shape, and it lands through executeEntryToBattlefieldLocked —
// the shared finisher every RESUMABLE entry crosses (a reanimation, a
// search, an exile return, stack resolution, a land play, a token).
//
// The pick is trimmed while a candidate remains and WITHDRAWN when the
// last one has been reanimated out from under it; before #1069 it kept
// offering a creature that was on the battlefield.
func TestAGraveyardPickIsPrunedWhenItsCandidateIsReanimated(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Regrowth")
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	first := permanentInGraveyard(t, g, me, "Buried One")
	second := permanentInGraveyard(t, g, me, "Buried Two")

	id, ran := queueGraveyardPick(t, g, me.ID, source, []uuid.UUID{first, second})

	// Somebody else's effect reanimates one of the candidates while the
	// prompt is open.
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(first, ZoneBattlefield, me.ID); err != nil {
			t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
		}
	})
	if !g.Battlefield.Contains(first) {
		t.Fatal("setup: the reanimated card is not on the battlefield")
	}
	c := findChoice(g, id)
	if c == nil {
		t.Fatal("one candidate is still in the graveyard; the prompt must not be withdrawn")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != second {
		t.Errorf("candidates after the prune = %v, want only the card still in the graveyard — "+
			"a card that left FOR THE BATTLEFIELD left the pick's zone like any other", c.ChooseCards)
	}

	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(second, ZoneBattlefield, me.ID); err != nil {
			t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
		}
	})
	if findChoice(g, id) != nil {
		t.Error("the prompt survives a graveyard with none of its candidates in it")
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, me.ID) == nil {
		t.Error("the withdrawal is not in the log; a stall dump has to say where the prompt went")
	}
	// A pick that is no run's leg has no continuation to settle — the
	// departure table's existing answer for the kind, unwidened here.
	if *ran != 0 {
		t.Errorf("the withdrawn pick ran its continuation %d times", *ran)
	}
	assertTableIsFree(t, g)
}

// TestADiscardPromptIsWithdrawnWhenItsHandIsPutOntoTheBattlefield is the
// issue's second shape — the Warp World door — and it lands through
// putOntoBattlefieldFromZoneLocked, the hand / library batch that is
// deliberately not resumable and so cannot share the finisher above.
//
// The prompt is one leg of a discard RUN, so the withdrawal has to
// settle that leg with "nothing discarded" and let the rest of the
// printed instruction happen (#1016's dropDefault, ADR 0013 §5y).
func TestADiscardPromptIsWithdrawnWhenItsHandIsPutOntoTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Mind Rot")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := permanentInHand(t, g, me, "Warped Bear")

	run := discardRunOverOneSeat(t, g, DiscardPrompt{Player: me.ID, Source: source, N: 1})
	if discardPromptFor(g, me.ID) == nil {
		t.Fatal("setup: no prompt over the one-card hand")
	}

	g.WithWriteLock(func() {
		entered, err := g.PutFromHandOntoBattlefieldForEffect(card, HandEntryOptions{Controller: me.ID})
		if err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
		if entered == uuid.Nil {
			t.Fatal("setup: nothing entered the battlefield")
		}
	})

	if !g.Battlefield.Contains(card) {
		t.Fatal("setup: the card is not on the battlefield")
	}
	if c := discardPromptFor(g, me.ID); c != nil {
		t.Fatalf("the prompt survives with %d candidates and none of them in the hand — "+
			"no answer it would accept exists", len(c.ChooseCards))
	}
	if run.ran != 1 {
		t.Fatalf("the run's continuation ran %d times, want 1 — a withdrawn leg settles with "+
			"nothing discarded, it does not strand the instruction", run.ran)
	}
	if run.got.Discarded(me.ID) {
		t.Errorf("a card put onto the battlefield was not DISCARDED: %v", run.got)
	}
	if len(run.got) != 1 {
		t.Errorf("the seat was ASKED, so it has an entry: %v", run.got)
	}
	assertTableIsFree(t, g)
}

// TestASandboxMoveOntoTheBattlefieldPrunesTheSameWay is the third
// landing, moveCardByRefLocked's inline entry branch: the admin
// move_card into the battlefield or the stack, which reaches neither of
// the two finishers above.
func TestASandboxMoveOntoTheBattlefieldPrunesTheSameWay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Regrowth")
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	buried := permanentInGraveyard(t, g, me, "Buried One")

	id, _ := queueGraveyardPick(t, g, me.ID, source, []uuid.UUID{buried})

	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneGraveyard, Owner: me.ID},
		ZoneRef{Kind: ZoneBattlefield},
		buried,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}

	if !g.Battlefield.Contains(buried) {
		t.Fatal("setup: the card is not on the battlefield")
	}
	if findChoice(g, id) != nil {
		t.Error("the prompt survives a graveyard whose only candidate is on the battlefield")
	}
	assertTableIsFree(t, g)
}

// TestAnEntryCannotRunTheExitPrimitivesPruneToo is the "no double-run"
// half of #1069, and it is a structural fact rather than a count: the
// exit primitive's prune block cannot fire for an entry, because
// routeDestinationLocked refuses the two destinations the entry path
// owns. So the new call at the entry landings adds a door, it does not
// double one.
func TestAnEntryCannotRunTheExitPrimitivesPruneToo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := permanentInGraveyard(t, g, me, "Buried One")

	for _, dst := range []ZoneKind{ZoneBattlefield, ZoneStack} {
		g.WithWriteLock(func() {
			zone, _, err := g.routeDestinationLocked(card, dst, uuid.Nil)
			if !errors.Is(err, ErrZoneNotFound) || zone != nil {
				t.Errorf("routeDestinationLocked(%s) = (%v, %v), want (nil, ErrZoneNotFound) — "+
					"an entry is not an exit, and the exit primitive must not route one", dst, zone, err)
			}
		})
	}
}
