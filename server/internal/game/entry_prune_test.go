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

// --- #1175: the second prune the entry funnel owes -------------------
//
// pruneStaleZoneChangeChoicesLocked drops a queued prompt asking about
// a move of a card OUT of a zone the card has since left (#605). An
// ENTRY empties a hand slot or lifts a card out of a graveyard exactly
// as an exit does, and #1069 took only the card-set prune through the
// new door. The two shapes below are the gap and its idempotency.

// stalableGraveyardExit starts a CR 903.9-pausing exile of a commander
// sitting in `owner`'s GRAVEYARD and returns the route's continuation
// counter. OldZone is the graveyard rather than the battlefield, which
// is what makes an arrival ON the battlefield able to strand it:
// pausedZoneChangeStaleLocked's last line asks whether the card is
// still in the zone the paused move would leave.
func stalableGraveyardExit(t *testing.T, g *Game, owner *Player, cardID uuid.UUID) *int {
	t.Helper()
	ran := 0
	g.mu.Lock()
	paused, err := g.routeCardToZoneLocked(zoneRoute{
		CardID: cardID,
		Dst:    ZoneExile,
		Actor:  owner.ID,
		then:   func(*Game) error { ran++; return nil },
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("routeCardToZoneLocked: %v", err)
	}
	if !paused {
		t.Fatal("setup: the CR 903.9 prompt did not pause the exile out of the graveyard")
	}
	if ran != 0 {
		t.Fatalf("setup: a paused route is not terminal, but its continuation ran %d times", ran)
	}
	return &ran
}

// TestAnArrivalPrunesAStaleMoveOutOfTheGraveyard is #1175's shape: a
// commander's CR 903.9 prompt is open over a move OUT of its owner's
// graveyard, and the card is reanimated while the question hangs. The
// exile it asks about can never happen — the card is not in the
// graveyard any more — so the prompt is as stale as one whose card was
// exiled by somebody else, and before #1175 only the exile ran the
// prune.
//
// The withdrawal is terminal for the route it takes away (#865), so
// the route's continuation runs: a batch sequenced through it carries
// on rather than stalling behind a question nobody can answer.
func TestAnArrivalPrunesAStaleMoveOutOfTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	g.WithWriteLock(func() { owner.Graveyard.Cards = nil })
	cmdID := seatCommander(t, owner.Graveyard, owner)

	ran := stalableGraveyardExit(t, g, owner, cmdID)
	prompt := expectCommanderPrompt(t, g, owner)

	// The arrival: somebody reanimates the commander under the open
	// question. This is the door #1069 opened and #1175 widens.
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(cmdID, ZoneBattlefield, owner.ID); err != nil {
			t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
		}
	})

	if !g.Battlefield.Contains(cmdID) {
		t.Fatal("setup: the reanimated commander is not on the battlefield")
	}
	if findChoice(g, prompt.ID) != nil {
		t.Error(`the CR 903.9 prompt survives a graveyard its card has left FOR THE BATTLEFIELD.
An entry empties a graveyard slot exactly as an exit does, and the exile the prompt
asks about can never happen — pruneStaleZoneChangeChoicesLocked has to run at the
entry door too (#1175).`)
	}
	if *ran != 1 {
		t.Errorf("the abandoned route's continuation ran %d times, want 1 — a withdrawal is "+
			"terminal for the route it takes away (#865)", *ran)
	}
	assertTableIsFree(t, g)
}

// TestTheArrivalFunnelIsIdempotent is the other half of "no double-run".
// The structural fact that an entry cannot reach the exit primitive's
// prune block is pinned by the test below; this pins the hazard the
// stale-move prune brings that the card-set prune did not — its
// terminal outcome is a ROUTE continuation (abandonZoneRouteLocked),
// so a funnel that ran twice over one arrival would run somebody's
// printed instruction twice.
//
// It cannot: the prune re-reads the live queue, and the withdrawn
// prompt is no longer in it to be found a second time.
func TestTheArrivalFunnelIsIdempotent(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	g.WithWriteLock(func() { owner.Graveyard.Cards = nil })
	cmdID := seatCommander(t, owner.Graveyard, owner)

	ran := stalableGraveyardExit(t, g, owner, cmdID)
	expectCommanderPrompt(t, g, owner)

	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(cmdID, ZoneBattlefield, owner.ID); err != nil {
			t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
		}
		// Two more calls by hand, standing in for a landing that
		// funnelled twice: neither may re-abandon a route already gone.
		g.pruneChoicesAfterArrivalLocked()
		g.pruneChoicesAfterArrivalLocked()
	})

	if *ran != 1 {
		t.Errorf("the abandoned route's continuation ran %d times over three funnel calls, want 1", *ran)
	}
	assertTableIsFree(t, g)
}

// TestAnEntryCannotRunTheExitPrimitivesPruneToo is the "no double-run"
// half of #1069 — and of #1175, which added the second prune to the
// same funnel. It is a structural fact rather than a count: the exit
// primitive's prune block cannot fire for an entry, because
// routeDestinationLocked refuses the two destinations the entry path
// owns. So the new calls at the entry landings add a door, they do not
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
