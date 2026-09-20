package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// player_counters_test.go is ADR 0056 PR 1's half of the test plan:
// items 12 (player counters through the window), 15 (the counter resume
// with a vanished target), 18 (the layer bump) and 19 (the proliferate
// placer), plus the placer riding onto the events that carry it.
//
// The damage-result items — infect, wither and toxic — are PR 2's, and
// their tests are the skipped block in infect_wither_toxic_test.go.

// --- helpers -------------------------------------------------------

// poisonOf reads a seat's poison count off the map, which is the
// authority; Player.Poison is the legacy mirror and is asserted
// separately where it matters.
func poisonOf(g *Game, playerID uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p.ID == playerID {
				n = p.Counters[CounterPoison]
			}
		}
	})
	return n
}

// legacyPoisonOf reads the pre-S13.2 int field.
func legacyPoisonOf(g *Game, playerID uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p.ID == playerID {
				n = p.Poison
			}
		}
	})
	return n
}

// playerCounterEvents returns every EventPlayerCounterPlaced in the log.
func playerCounterEvents(g *Game) []Event {
	var out []Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventPlayerCounterPlaced {
				out = append(out, ev)
			}
		}
	})
	return out
}

// onlyPlayerCounterEvent returns the single EventPlayerCounterPlaced in
// the log, failing when there is not exactly one. "Exactly one" is most
// of what ADR 0056 Decision 4 is about: infect and toxic from the same
// source are ONE placement, not two, because a halving replacement
// rounds each event separately.
func onlyPlayerCounterEvent(t *testing.T, g *Game) Event {
	t.Helper()
	evs := playerCounterEvents(g)
	if len(evs) != 1 {
		t.Fatalf("%d EventPlayerCounterPlaced in the log, want exactly 1: %+v", len(evs), evs)
	}
	return evs[0]
}

// doublePlayerCounters is a Vorinclex-shaped test replacement for the
// PLAYER half of RepEventCounter: it doubles any placement on `on`.
func doublePlayerCounters(on uuid.UUID, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterPlayer == on && ev.CounterDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta *= 2
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return on },
		Label:      label,
	}
}

// oneMorePlayerCounter is the Lae'zel-shaped sibling, so a board with
// both makes the window a genuine CR 616 ordering question rather than
// #792's identical-window skip: [double, +1] and [+1, double] differ.
func oneMorePlayerCounter(on uuid.UUID, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterPlayer == on && ev.CounterDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta++
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return on },
		Label:      label,
	}
}

// addCardCounters is a plus-one-counter test replacement on a CARD,
// used to prove the two halves of the kind stay apart.
func addCardCounters(on uuid.UUID, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterTarget == on && ev.CounterDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta++
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return uuid.Nil },
		Label:      label,
	}
}

// --- item 12: player counters go through the CR 614 window ----------

// TestPlayerCountersGoThroughTheReplacementWindow is ADR 0056 test plan
// item 12. Before this change a poison counter was a bare map write:
// AddPlayerCounterForEffect adjusted Player.Counters and returned, so a
// doubler had nothing to see.
func TestPlayerCountersGoThroughTheReplacementWindow(t *testing.T) {
	g := newActiveGame(t)
	victim := g.Seats[0]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(doublePlayerCounters(victim.ID, "double poison"))
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 3); err != nil {
			t.Fatalf("AddPlayerCounterForEffect: %v", err)
		}
	})

	if got := poisonOf(g, victim.ID); got != 6 {
		t.Errorf("poison = %d, want 6 (3 doubled)", got)
	}
	if got := legacyPoisonOf(g, victim.ID); got != 6 {
		t.Errorf("legacy Player.Poison = %d, want 6 — the mirror must follow the REPLACED delta, not the asked-for one", got)
	}
	ev := onlyPlayerCounterEvent(t, g)
	if ev.Target != victim.ID || ev.Label != CounterPoison || ev.Amount != 6 {
		t.Errorf("event = {target %v, label %q, amount %d}, want {%v, %q, 6}",
			ev.Target, ev.Label, ev.Amount, victim.ID, CounterPoison)
	}
}

// TestAPlayerCounterEventCarriesTheDeltaNotTheTotal is the other half
// of the event's contract, and the one that makes it a different KIND
// from EventCounterPlaced. The card event carries the post-change
// total, which is exactly what forced three catalog helpers to walk the
// log backwards to recover how many counters had been placed.
func TestAPlayerCounterEventCarriesTheDeltaNotTheTotal(t *testing.T) {
	g := newActiveGame(t)
	victim := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 4); err != nil {
			t.Fatalf("first: %v", err)
		}
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 3); err != nil {
			t.Fatalf("second: %v", err)
		}
	})

	evs := playerCounterEvents(g)
	if len(evs) != 2 {
		t.Fatalf("%d events, want 2", len(evs))
	}
	if evs[0].Amount != 4 || evs[1].Amount != 3 {
		t.Errorf("amounts = %d, %d; want 4, 3 (the DELTA each time, not the running total 4, 7)",
			evs[0].Amount, evs[1].Amount)
	}
	if got := poisonOf(g, victim.ID); got != 7 {
		t.Errorf("poison = %d, want 7", got)
	}
}

// TestAClampedPlayerCounterRemovalReportsWhatLanded pins the clamp. A
// removal of two from a player with one is a change of ONE, and a
// removal from a player with none is not a change at all — so it emits
// nothing, logs nothing and does not bump the layer version.
func TestAClampedPlayerCounterRemovalReportsWhatLanded(t *testing.T) {
	g := newActiveGame(t)
	victim := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 1); err != nil {
			t.Fatalf("seed: %v", err)
		}
	})

	before := readLayerVersion(g)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, -2); err != nil {
			t.Fatalf("clamped removal: %v", err)
		}
	})
	evs := playerCounterEvents(g)
	if len(evs) != 2 {
		t.Fatalf("%d events, want 2", len(evs))
	}
	if evs[1].Amount != -1 {
		t.Errorf("clamped removal reported %d, want -1 (one counter actually came off)", evs[1].Amount)
	}
	if got := poisonOf(g, victim.ID); got != 0 {
		t.Errorf("poison = %d, want 0", got)
	}
	if readLayerVersion(g) <= before {
		t.Errorf("a removal that landed did not bump the layer version")
	}

	// And now the no-op.
	before = readLayerVersion(g)
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, -1); err != nil {
			t.Fatalf("no-op removal: %v", err)
		}
	})
	if got := len(playerCounterEvents(g)); got != 2 {
		t.Errorf("%d events after a removal that moved nothing, want still 2", got)
	}
	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion moved from %d to %d on a placement that changed nothing", before, got)
	}
}

// TestAPlayerCounterEventIsNotACardCounterEvent is the reason ADR 0056
// Decision 5 minted a second kind. A player UUID carrying the label
// "poison" must never reach a card-counter trigger helper, and those
// helpers filter on Kind.
func TestAPlayerCounterEventIsNotACardCounterEvent(t *testing.T) {
	g := newActiveGame(t)
	victim := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 2); err != nil {
			t.Fatalf("AddPlayerCounterForEffect: %v", err)
		}
	})
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventCounterPlaced {
				t.Errorf("a player counter emitted EventCounterPlaced (target %v, label %q) — card-counter helpers compare Target against card IDs",
					ev.Target, ev.Label)
			}
		}
	})
}

// TestACardCounterReplacementDoesNotSeeAPlayerEvent is the safety
// property ADR 0056 Decision 5 relies on to keep ONE event kind for
// both targets: every counter replacement written before this change
// opens with a card lookup, which fails for uuid.Nil.
func TestACardCounterReplacementDoesNotSeeAPlayerEvent(t *testing.T) {
	g := newActiveGame(t)
	victim := g.Seats[0]
	bear := pushTypedTestCard(g, Card{
		Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: victim.ID, Controller: victim.ID,
	})
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(addCardCounters(bear, "card-only +1"))
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 2); err != nil {
			t.Fatalf("AddPlayerCounterForEffect: %v", err)
		}
	})
	if got := poisonOf(g, victim.ID); got != 2 {
		t.Errorf("poison = %d, want 2 — a CARD replacement applied to a PLAYER placement", got)
	}
}

// --- the CR 616.1 affected player ----------------------------------

// TestThePlayerGettingTheCountersOrdersTheReplacements is
// affectedPlayerForEvent's new arm. Two replacements apply, both
// controlled by somebody who is NOT the player receiving the counters,
// so the pre-ADR fallback ("the first gathered effect's controller")
// would put the prompt to the wrong seat — the Vorinclex-beside-a-
// doubler board.
func TestThePlayerGettingTheCountersOrdersTheReplacements(t *testing.T) {
	g := newActiveGame(t)
	victim, other := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(doublePlayerCounters(victim.ID, "their doubler"))
		g.RegisterReplacementForTest(oneMorePlayerCounter(victim.ID, "their plus one"))
		// Both effects answer `other` from their Controller hook, which
		// is what the fallback would read.
	})
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 2); err != nil {
			t.Fatalf("AddPlayerCounterForEffect: %v", err)
		}
	})
	_ = other

	p := onlyPrompt(t, g)
	if p.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %v, want %v", p.Kind, PendingChoiceReplacementOrder)
	}
	if p.Chooser != victim.ID {
		t.Errorf("chooser = %v, want %v (CR 616.1: the affected PLAYER orders them)", p.Chooser, victim.ID)
	}
	if got := poisonOf(g, victim.ID); got != 0 {
		t.Errorf("poison = %d while the prompt is open, want 0 — a paused placement must place nothing", got)
	}

	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, p.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if got := poisonOf(g, victim.ID); got != 5 && got != 6 {
		t.Errorf("poison = %d, want 5 ((2+1)*2 is 6, 2*2+1 is 5) — one of the two CR 616 orders", got)
	}
}

// --- item 15: the counter resume tolerates a vanished target --------

// TestTheCounterResumeSurvivesAVanishedCard is ADR 0056 test plan item
// 15 and the pre-existing bug Decision 3 names. The RepEventCounter
// resume used to `return g.applyCounterLocked(...)`, so a target that
// left between the prompt and the answer failed the ACTION — after the
// choice had been dequeued, which takes the prompt away with nothing to
// show for it. That is the #694 shape the life and damage arms already
// fixed.
func TestTheCounterResumeSurvivesAVanishedCard(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	// A TOKEN, because "the target left" has to mean left THE GAME:
	// applyCounterLocked finds a card in any zone, and a real card that
	// died is still in the graveyard taking counters (which is what
	// AddCounter's "the card may be in any zone" promises). CR 704.5d
	// is the one thing that makes a permanent stop existing, and a
	// -1/-1 counter landing on a creature token that has already died
	// is exactly the board ADR 0056's damage results produce.
	token := pushTypedTestCard(g, Card{
		// "Token" in the printed type line is what Card.IsToken reads.
		Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1,
		Owner: owner.ID, Controller: owner.ID,
	})
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(addCardCounters(token, "plus one"))
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCounter && ev.CounterTarget == token && ev.CounterDelta > 0
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner.ID },
			Label:      "double",
		})
	})
	if err := g.AddCounter(token, CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	p := onlyPrompt(t, g)

	// The token dies and CR 704.5d removes it from the game entirely
	// while the prompt is open.
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID}, token); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if cardExistsForTest(g, token) {
		t.Fatal("the token still exists after the CR 704.5d sweep — the test is not testing what it says")
	}

	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, p.ReplacementEffectIDs); err != nil {
		t.Fatalf(`answering the prompt returned %v.

The target left between the prompt and the answer. The counters simply
do not land — but the choice is already dequeued, so returning the error
fails the player's action AND takes their prompt away with nothing to
show for it (#694).`, err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts still open after the answer", len(g.PendingChoices))
	}
	found := false
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventEffectError {
				found = true
			}
		}
	})
	if !found {
		t.Error("no EventEffectError breadcrumb for the dropped counter placement")
	}
}

// TestAnEliminatedPlayerTakesNoCounters is the player half of "the
// target is gone".
//
// It is NOT reachable through the CR 616 resume the way the card half
// is: eliminating a seat DROPS its pending choices, so a prompt about
// counters going on a player who then leaves is gone before anybody can
// answer it. The reachable shape is the unpaused one — an effect
// placing counters on a player eliminated earlier in the same
// resolution, which is ordinary in a batch ("each opponent gets a
// poison counter") when an SBA took one of them out partway through.
//
// An eliminated seat is never removed from g.Seats, so a nil check
// alone would let it happen. That is #808's bug on the life side, and
// applyResolvedLifeChangeLocked's guard is the one this mirrors. The
// resume's own tolerance stays as a defence in depth: it is the arm
// that has to not fail an action whose prompt is already dequeued.
func TestAnEliminatedPlayerTakesNoCounters(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	victim := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 2); err != nil {
			t.Fatalf("seed: %v", err)
		}
	})
	if err := g.Concede(victim.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	var err error
	g.WithWriteLock(func() {
		err = g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 5)
	})
	if !errors.Is(err, ErrPlayerEliminated) {
		t.Errorf("placing counters on an eliminated player returned %v, want ErrPlayerEliminated", err)
	}
	if got := poisonOf(g, victim.ID); got != 2 {
		t.Errorf("poison = %d, want 2 — an eliminated player took counters (CR 800.4a)", got)
	}
}

// TestEliminationDropsAnOpenCounterPrompt is the other half of the
// statement above, and the reason the resume's ErrPlayerEliminated arm
// is defence in depth rather than the live path: the prompt goes with
// the player.
func TestEliminationDropsAnOpenCounterPrompt(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	victim := g.Seats[0]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(doublePlayerCounters(victim.ID, "double"))
		g.RegisterReplacementForTest(oneMorePlayerCounter(victim.ID, "plus one"))
	})
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(victim.ID, CounterPoison, 2); err != nil {
			t.Fatalf("AddPlayerCounterForEffect: %v", err)
		}
	})
	if p := onlyPrompt(t, g); p.Chooser != victim.ID {
		t.Fatalf("chooser = %v, want %v", p.Chooser, victim.ID)
	}

	if err := g.Concede(victim.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts still open after the chooser was eliminated", len(g.PendingChoices))
	}
	if got := poisonOf(g, victim.ID); got != 0 {
		t.Errorf("poison = %d, want 0 — the paused placement must not land", got)
	}
}

// TestTheCounterResumeSweepsStateBasedActions is the other half of
// Decision 3's hardening, and the half ADR 0056's damage results make
// routine: a -1/-1 placement that PAUSED has no resolution bookend to
// fall back on, so the creature it brings to zero toughness has to die
// at the answer.
func TestTheCounterResumeSweepsStateBasedActions(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	bear := pushTypedTestCard(g, Card{
		Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: owner.ID, Controller: owner.ID,
	})
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(addCardCounters(bear, "plus one"))
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCounter && ev.CounterTarget == bear && ev.CounterDelta > 0
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner.ID },
			Label:      "double",
		})
	})
	// One -1/-1 counter becomes at least two either way, which takes a
	// 2/2 to zero toughness (CR 704.5f).
	if err := g.AddCounter(bear, CounterMinusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	p := onlyPrompt(t, g)
	if !onBattlefieldForTest(g, bear) {
		t.Fatal("the bear died before the prompt was answered")
	}

	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, p.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if onBattlefieldForTest(g, bear) {
		t.Error(`the bear is still on the battlefield.

Answering a prompt is an action boundary, and the unpaused counter
paths get their sweep from the resolution bookend they run inside. A
placement that PAUSED has no bookend, so the CR 704.5f sweep has to
happen at the answer (ADR 0056 Decision 3).`)
	}
}

// cardExistsForTest reports whether the engine can still find the card
// in any tracked zone. CR 704.5d is the only thing that makes the
// answer no for a permanent that has moved.
func cardExistsForTest(g *Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		_, found = g.LookupCardForEffect(id)
	})
	return found
}

// onBattlefieldForTest reports whether the card is still on the
// battlefield.
func onBattlefieldForTest(g *Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

// --- item 18: the layer bump ---------------------------------------

const poisonCountCDAOracle = "poison-count-cda"

// poisonCountCDAForTest is Vishgraz's shape: a layer-7a CDA whose P/T
// is the number of poison counters the source's OPPONENTS have. The
// static reads a player counter, which is precisely the input nothing
// invalidated before ADR 0056 Decision 5.
func poisonCountCDAForTest() StaticAbility {
	return StaticAbility{
		Layer:    Layer7PT,
		SubLayer: SubLayer7A_CDA,
		AppliesTo: func(target *Card, _ *Game, source *Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *Characteristic, _ *Card, g *Game, source *Card) {
			n := 0
			for _, p := range g.Seats {
				if p.ID != source.Controller {
					n += p.Counters[CounterPoison]
				}
			}
			c.Power, c.Toughness = n, n
		},
	}
}

// TestAPoisonCountStaticUpdatesInTheSameAction is ADR 0056 test plan
// item 18, and the reason PR 1 exists ahead of the cards. A "corrupted"
// static or a poison-count P/T is a layer input; before this change a
// player counter was a bare map write, so the cached resolution
// survived the poison and the creature kept the wrong size until some
// unrelated permanent happened to move.
func TestAPoisonCountStaticUpdatesInTheSameAction(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == poisonCountCDAOracle {
			return []StaticAbility{poisonCountCDAForTest()}
		}
		return nil
	})
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	hive := pushTypedTestCard(g, Card{
		Name:     "Poison Reader",
		TypeLine: "Creature — Phyrexian Horror",
		OracleID: poisonCountCDAOracle,
		Owner:    me.ID, Controller: me.ID,
	})

	if got := layeredBattlefieldCard(t, g, hive).Effective().Power; got != 0 {
		t.Fatalf("baseline power = %d, want 0", got)
	}

	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(them.ID, CounterPoison, 3); err != nil {
			t.Fatalf("AddPlayerCounterForEffect: %v", err)
		}
	})

	if got := layeredBattlefieldCard(t, g, hive).Effective().Power; got != 3 {
		t.Errorf(`power = %d, want 3.

The static reads an opponent's poison count and the count just changed,
so the cached layer resolution is stale. EventPlayerCounterPlaced is
what invalidates it (ADR 0056 Decision 5).`, got)
	}
}

// TestTheSandboxPoisonVerbsBumpTheLayerVersion is the same property for
// the two manual verbs. They SKIP the replacement window on purpose —
// set_poison is set semantics and there is nothing for a replacement to
// say about "make the total 7" — but a static that reads the count must
// not care which way the count arrived.
func TestTheSandboxPoisonVerbsBumpTheLayerVersion(t *testing.T) {
	g := newActiveGame(t)
	victim := g.Seats[0]

	before := readLayerVersion(g)
	if err := g.SetPoison(victim.ID, 7); err != nil {
		t.Fatalf("SetPoison: %v", err)
	}
	if got := readLayerVersion(g); got <= before {
		t.Errorf("SetPoison did not bump the layer version: was %d, now %d", before, got)
	}
	ev := onlyPlayerCounterEvent(t, g)
	if ev.Amount != 7 {
		t.Errorf("SetPoison event amount = %d, want 7 (the DELTA from 0)", ev.Amount)
	}

	before = readLayerVersion(g)
	if err := g.AddPlayerCounter(victim.ID, CounterPoison, 2); err != nil {
		t.Fatalf("AddPlayerCounter: %v", err)
	}
	if got := readLayerVersion(g); got <= before {
		t.Errorf("AddPlayerCounter did not bump the layer version: was %d, now %d", before, got)
	}
	if evs := playerCounterEvents(g); len(evs) != 2 || evs[1].Amount != 2 {
		t.Errorf("AddPlayerCounter events = %+v, want a second one with amount 2", evs)
	}

	// A set to the value it already has is not a change.
	before = readLayerVersion(g)
	if err := g.SetPoison(victim.ID, 9); err != nil {
		t.Fatalf("SetPoison no-op: %v", err)
	}
	if got := readLayerVersion(g); got != before {
		t.Errorf("a SetPoison to the current total bumped the layer version: %d → %d", before, got)
	}
}

// --- item 19: the proliferate placer -------------------------------

// TestProliferateNamesTheProliferatingPlayerAsThePlacer is ADR 0056
// test plan item 19. CR 701.34 is "GIVE each another counter", so the
// proliferating player is who puts them — on permanents and on players
// alike. It is what makes Vorinclex halve an opponent's proliferate
// onto your creature rather than reading the target's controller.
func TestProliferateNamesTheProliferatingPlayerAsThePlacer(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := pushTypedTestCard(g, Card{
		Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(them.ID, CounterPoison, 1); err != nil {
			t.Fatalf("seed poison: %v", err)
		}
		if err := g.applyCounterLocked(bear, CounterPlusOne, 1); err != nil {
			t.Fatalf("seed counter: %v", err)
		}
	})

	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{bear}, []uuid.UUID{them.ID}); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})

	if got := poisonOf(g, them.ID); got != 2 {
		t.Errorf("poison after proliferate = %d, want 2", got)
	}

	var playerPlacement, cardPlacement Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			switch {
			case ev.Kind == EventPlayerCounterPlaced && ev.Target == them.ID:
				playerPlacement = ev
			case ev.Kind == EventCounterPlaced && ev.Target == bear && ev.Actor != uuid.Nil:
				cardPlacement = ev
			}
		}
	})
	if playerPlacement.Actor != me.ID {
		t.Errorf("player placement actor = %v, want the proliferating player %v", playerPlacement.Actor, me.ID)
	}
	if cardPlacement.Actor != me.ID {
		t.Errorf("card placement actor = %v, want the proliferating player %v", cardPlacement.Actor, me.ID)
	}
}

// TestAPlacementWithNoKnownPlacerCarriesNil pins the other direction.
// uuid.Nil means UNKNOWN, and a "you put" reader falls back to its own
// heuristic on it — which is weaker than printed rather than stronger,
// and is what keeps a paid cost or a sandbox edit reading the way it
// always has.
func TestAPlacementWithNoKnownPlacerCarriesNil(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	bear := pushTypedTestCard(g, Card{
		Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: owner.ID, Controller: owner.ID,
	})
	if err := g.AddCounter(bear, CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventCounterPlaced && ev.Target == bear && ev.Actor != uuid.Nil {
				t.Errorf("the sandbox AddCounter verb named %v as the placer; a hand-edit is nobody putting counters", ev.Actor)
			}
		}
	})
}
