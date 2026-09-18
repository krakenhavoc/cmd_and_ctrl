package game

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// duration_test.go pins the CR 611.2 duration model (ADR 0063, #755):
// each kind ends at exactly its own boundary and at no other, and the
// one expiry function is reached from all three sweep sites.
//
// Card-level behaviour (Act of Treason, Sower of Temptation, Mass
// Diminish) lives in the effects package.

// registerDurationMarker installs a harmless layer-6 grant with the
// given duration and returns nothing: these tests assert on the
// registry, which is what the duration decides.
func registerDurationMarker(g *Game, target uuid.UUID, d Duration, label string) {
	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer6Ability,
			AppliesTo: scopedPinnedTo(target),
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Abilities = append(c.Abilities, "vigilance")
			},
		}, uuid.New(), label, d)
	})
}

// advanceOneTurn walks the cursor until the active seat changes.
func advanceOneTurn(t *testing.T, g *Game) {
	t.Helper()
	start := g.Turn.ActiveSeat
	for i := 0; i < 40; i++ {
		if g.Turn.ActiveSeat != start {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("cursor never left seat %d's turn", start)
}

// TestStartCountsTheFirstTurn — the starting seat's turn is the one
// turn that does not come through the rotation seam, so Start has to
// stamp it or every "until your next turn" effect made on turn one
// would be off by a turn.
func TestStartCountsTheFirstTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if got := g.Seats[g.StartingSeat].TurnsBegun; got != 1 {
		t.Errorf("starting seat TurnsBegun = %d, want 1", got)
	}
	for i, p := range g.Seats {
		if i == g.StartingSeat && p.TurnsBegun != 1 {
			t.Errorf("seat %d TurnsBegun = %d, want 1", i, p.TurnsBegun)
		}
		if i != g.StartingSeat && p.TurnsBegun != 0 {
			t.Errorf("seat %d TurnsBegun = %d before its first turn, want 0", i, p.TurnsBegun)
		}
	}
}

// TestUntilYourNextTurnSurvivesEveryOpponentsTurn is the headline for
// #755: an effect that lasts until seat 0's next turn must sit through
// seats 1, 2 and 3 and end only when seat 0 comes back round. A
// duration expressed on Turn.Number could not tell those four turns
// apart — they share one round number.
func TestUntilYourNextTurnSurvivesEveryOpponentsTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	bear := pushScopedTestCreature(g, me.ID, 2, 2)

	var d Duration
	g.WithWriteLock(func() { d = g.UntilYourNextTurnDuration(me.ID) })
	registerDurationMarker(g, bear, d, "until your next turn")

	for seat := 1; seat <= 3; seat++ {
		advanceOneTurn(t, g)
		if g.Turn.ActiveSeat != seat {
			t.Fatalf("expected seat %d's turn, got %d", seat, g.Turn.ActiveSeat)
		}
		if n := len(g.ScopedStatics); n != 1 {
			t.Fatalf("the effect ended during seat %d's turn (%d entries left)", seat, n)
		}
	}

	advanceOneTurn(t, g)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("expected to be back on seat 0, got %d", g.Turn.ActiveSeat)
	}
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("the effect survived its own player's next turn (%d entries)", n)
	}
}

// TestUntilYourNextTurnEndsAsTheTurnBegins pins the boundary itself:
// the effect is gone before the untap step's turn-based action runs,
// because the turn has begun (CR 500.1) and the untapping happens
// inside it (CR 502.1).
func TestUntilYourNextTurnEndsAsTheTurnBegins(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushScopedTestCreature(g, me.ID, 2, 2)

	// Tap the bear so the untap step has real work to do; if the
	// sweep ran after untapping, the assertion below would still
	// pass, so the test also checks the untap happened at all.
	g.WithWriteLock(func() {
		if c, ok := g.battlefieldCardLocked(bear); ok {
			c.Tapped = true
		}
	})

	var d Duration
	g.WithWriteLock(func() { d = g.UntilYourNextTurnDuration(me.ID) })
	registerDurationMarker(g, bear, d, "until your next turn")

	advanceOneTurn(t, g) // opponent
	if n := len(g.ScopedStatics); n != 1 {
		t.Fatalf("effect ended on the opponent's turn (%d entries)", n)
	}
	advanceOneTurn(t, g) // back to me
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("effect survived the start of its player's turn (%d entries)", n)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bear && c.Tapped {
				t.Error("the untap step did not run; the boundary assertion is vacuous")
			}
		}
	})
}

// TestUntilYourNextTurnMadeOnYourOwnTurnLastsAFullRound — the stamp is
// `TurnsBegun + 1`, so an effect created during your own turn does not
// end at the next turn boundary. It has to wait for YOUR next turn.
func TestUntilYourNextTurnMadeOnYourOwnTurnLastsAFullRound(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	bear := pushScopedTestCreature(g, me.ID, 2, 2)

	var d Duration
	g.WithWriteLock(func() { d = g.UntilYourNextTurnDuration(me.ID) })
	if d.ExpiresAtTurnsBegun != 2 {
		t.Fatalf("ExpiresAtTurnsBegun = %d on the holder's own first turn, want 2",
			d.ExpiresAtTurnsBegun)
	}
	registerDurationMarker(g, bear, d, "until your next turn")

	advanceOneTurn(t, g)
	if n := len(g.ScopedStatics); n != 1 {
		t.Fatalf("an effect made on its own player's turn ended at the very next turn boundary")
	}
}

// TestUntilYourNextTurnEndsWhenADepartedPlayersTurnWouldHaveBegun is
// CR 800.4m: the rotation steps over an eliminated seat, and that
// never-taken turn still counts, so an effect keyed to it ends at the
// right moment rather than lasting for the rest of the game.
func TestUntilYourNextTurnEndsWhenADepartedPlayersTurnWouldHaveBegun(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, leaver := g.Seats[0], g.Seats[2]
	bear := pushScopedTestCreature(g, me.ID, 2, 2)

	var d Duration
	g.WithWriteLock(func() { d = g.UntilYourNextTurnDuration(leaver.ID) })
	registerDurationMarker(g, bear, d, "until the leaver's next turn")

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := len(g.ScopedStatics); n != 1 {
		t.Fatalf("conceding ended the effect immediately; CR 800.4m says it waits")
	}

	advanceOneTurn(t, g) // seat 1
	if n := len(g.ScopedStatics); n != 1 {
		t.Fatalf("the effect ended on seat 1's turn (%d entries)", n)
	}
	// Seat 2 has left, so the cursor steps over it to seat 3 — and
	// seat 2's never-taken turn is what ends the effect.
	advanceOneTurn(t, g)
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("the departed player's turn came and went and the effect is still live (%d entries)", n)
	}
}

// TestForAsLongAsEndsWhenTheSourceLeaves — CR 611.2b. The condition is
// re-evaluated on every layer pass, and a zone change is a layer
// version bump, so the effect ends the moment the source does.
func TestForAsLongAsEndsWhenTheSourceLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := pushScopedTestCreature(g, me.ID, 1, 1)
	bear := pushScopedTestCreature(g, me.ID, 2, 2)

	var d Duration
	var ok bool
	g.WithWriteLock(func() { d, ok = g.ForAsLongAsOnBattlefieldDuration(source) })
	if !ok {
		t.Fatal("ForAsLongAsOnBattlefieldDuration refused a source that is on the battlefield")
	}
	registerDurationMarker(g, bear, d, "for as long as the source remains")

	g.ReadSnapshot(func() {})
	if n := len(g.ScopedStatics); n != 1 {
		t.Fatalf("the effect ended while its source was still there (%d entries)", n)
	}

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, source); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: source,
			OldZone: ZoneBattlefield, NewZone: ZoneGraveyard})
	})
	g.ReadSnapshot(func() {})
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("the effect outlived its source (%d entries)", n)
	}
}

// TestForAsLongAsEndsWhenTheSourceIsFlickered is CR 400.7: the
// permanent that comes back is a NEW object, so "for as long as THIS
// creature remains" is about something that no longer exists. The
// battlefield-entry stamp is what lets the engine tell, and dropping
// it from the condition is the bug this test fails on.
func TestForAsLongAsEndsWhenTheSourceIsFlickered(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := pushScopedTestCreature(g, me.ID, 1, 1)
	bear := pushScopedTestCreature(g, me.ID, 2, 2)

	var d Duration
	g.WithWriteLock(func() { d, _ = g.ForAsLongAsOnBattlefieldDuration(source) })
	registerDurationMarker(g, bear, d, "for as long as the source remains")

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, g.Exile, source); err != nil {
			t.Fatalf("MoveCard out: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: source,
			OldZone: ZoneBattlefield, NewZone: ZoneExile})
		if _, err := MoveCard(g.Exile, g.Battlefield, source); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: source,
			OldZone: ZoneExile, NewZone: ZoneBattlefield})
	})
	g.ReadSnapshot(func() {})

	if _, ok := g.battlefieldCardLocked(source); !ok {
		t.Fatal("setup: the source did not come back")
	}
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("a flickered source kept the effect alive (%d entries) — CR 400.7 says it is a new object", n)
	}
}

// TestForAsLongAsNeverStartsWhenTheConditionIsAlreadyFalse — CR 611.2b.
// The constructor refuses rather than registering an effect that would
// die on its first recompute.
func TestForAsLongAsNeverStartsWhenTheConditionIsAlreadyFalse(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if _, ok := g.ForAsLongAsOnBattlefieldDuration(uuid.New()); ok {
			t.Error("a source that is not on the battlefield produced a live duration")
		}
	})
}

// TestForAsLongAsYouControlEndsWhenControlOfTheSourceChanges — the
// other CR 611.2b condition. The source is still on the battlefield;
// what stopped is the "you control" half.
func TestForAsLongAsYouControlEndsWhenControlOfTheSourceChanges(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := pushScopedTestCreature(g, me.ID, 1, 1)
	bear := pushScopedTestCreature(g, me.ID, 2, 2)

	var d Duration
	var ok bool
	g.WithWriteLock(func() { d, ok = g.ForAsLongAsYouControlDuration(source, me.ID) })
	if !ok {
		t.Fatal("the constructor refused a source its player controls")
	}
	registerDurationMarker(g, bear, d, "for as long as you control the source")
	g.ReadSnapshot(func() {})
	if n := len(g.ScopedStatics); n != 1 {
		t.Fatalf("the effect ended while its player still controlled the source (%d entries)", n)
	}

	g.WithWriteLock(func() {
		if c, ok := g.battlefieldCardLocked(source); ok {
			c.Controller = opp.ID
			c.BaseController = opp.ID
		}
		g.layerVersion.Add(1)
	})
	g.ReadSnapshot(func() {})
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("losing control of the source left the effect live (%d entries)", n)
	}
}

// TestIndefiniteSurvivesManyTurns — CR 611.2a. No cleanup step and no
// turn boundary ends an effect with no stated duration.
func TestIndefiniteSurvivesManyTurns(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerDurationMarker(g, bear, IndefiniteDuration(), "no stated duration")

	for i := 0; i < 9; i++ {
		advanceOneTurn(t, g)
		if n := len(g.ScopedStatics); n != 1 {
			t.Fatalf("an effect with no stated duration ended after %d turns", i+1)
		}
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != bear {
				continue
			}
			if !eotHasAbilityForTest(c.Effective().Abilities, "vigilance") {
				t.Error("the effect is in the registry but is not applying")
			}
		}
	})
}

// TestUntilEndOfTurnStillEndsAtCleanup is the regression guard for the
// duration rewrite: the one duration that already worked must keep
// working, and must not be ended early by either of the two new sweep
// sites.
func TestUntilEndOfTurnStillEndsAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	var d Duration
	g.WithWriteLock(func() { d = g.UntilEndOfTurnDuration() })
	registerDurationMarker(g, bear, d, "until end of turn")

	// Several recomputes mid-turn must not touch it.
	for i := 0; i < 3; i++ {
		g.ReadSnapshot(func() {})
		if n := len(g.ScopedStatics); n != 1 {
			t.Fatalf("the layer-pass sweep ended an until-end-of-turn effect mid-turn")
		}
	}
	advancePastScopedCleanup(t, g)
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("until-end-of-turn survived the cleanup step (%d entries)", n)
	}
}

// TestPinnedDurationEndsWhenItsObjectLeaves is the registry's garbage
// collection (ADR 0063 Decision 6): an indefinite effect that exists
// to move one permanent has nothing left to do once that permanent
// does not exist, and leaving it in the registry would keep the game
// off the full-restore path for the rest of the game.
func TestPinnedDurationEndsWhenItsObjectLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushScopedTestCreature(g, me.ID, 2, 2)

	var d Duration
	g.WithWriteLock(func() { d = g.PinnedTo(IndefiniteDuration(), bear) })
	if d.Pinned != bear || d.PinnedEnteredAt == 0 {
		t.Fatalf("PinnedTo did not capture the object: %+v", d)
	}
	registerDurationMarker(g, bear, d, "pinned, no stated duration")

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, bear); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: bear,
			OldZone: ZoneBattlefield, NewZone: ZoneGraveyard})
	})
	g.ReadSnapshot(func() {})
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("a pinned effect outlived the object it was pinned to (%d entries)", n)
	}
}

// TestCensusLabelsNameTheDuration — a refused restore point has to
// tell an operator what is holding it up, and "a scoped static" is not
// an answer when the entry can now last the whole game.
func TestCensusLabelsNameTheDuration(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerDurationMarker(g, bear, IndefiniteDuration(), "Agent of Treachery — gain control")

	snap := g.CaptureSnapshot()
	if snap.Continuations.ScopedStatics != 1 {
		t.Fatalf("census counted %d scoped statics, want 1", snap.Continuations.ScopedStatics)
	}
	found := false
	for _, l := range snap.Continuations.Labels {
		if containsAll(l, "no stated duration", "Agent of Treachery") {
			found = true
		}
	}
	if !found {
		t.Errorf("census labels %v name neither the duration nor the effect", snap.Continuations.Labels)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

func eotHasAbilityForTest(abilities []string, kw string) bool {
	for _, a := range abilities {
		if a == kw {
			return true
		}
	}
	return false
}
