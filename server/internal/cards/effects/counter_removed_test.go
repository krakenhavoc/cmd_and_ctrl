package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_removed_test.go — #2466: "whenever a [kind] counter is removed
// from ~" triggers once per counter (CR 603.2c, Protean Hydra's ruling),
// by any cause, and not for counters that vanish because the permanent
// left the battlefield (CR 122.2).

const (
	oracleProteanHydra     = "6e32b958-70d3-4f3d-b10e-d6c8cb27dd93"
	oracleDinosaursSpaceSh = "e3b4314b-e7a2-449a-af5a-817c8260adf1"
)

func crRemove(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	putCounters(t, g, id, kind, -n)
}

// crEndOfTurn lets everything already on the stack resolve, walks to
// the end step and lets whatever fires there resolve.
func crEndOfTurn(t *testing.T, g *game.Game) {
	t.Helper()
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
}

func crHydra(t *testing.T, g *game.Game, counters int) uuid.UUID {
	t.Helper()
	return pr8cPermanent(t, g, g.Seats[0].ID, "Protean Hydra", oracleProteanHydra, 0, 0, counters)
}

func TestProteanHydraOneCounterRemovedIsOneTrigger(t *testing.T) {
	g := newCatalogGame(t)
	id := crHydra(t, g, 4)
	crRemove(t, g, id, game.CounterPlusOne, 1)
	if g.Stack.Size()+len(g.PendingTriggers) != 1 {
		t.Fatalf("%d stack items after one counter was removed, want 1", g.Stack.Size())
	}
	crEndOfTurn(t, g)
	if got := pr8Counters(g, id, game.CounterPlusOne); got != 3+2 {
		t.Fatalf("counters: got %d, want 5 (4, one removed, two back at end step)", got)
	}
}

func TestProteanHydraThreeCountersRemovedAreThreeTriggers(t *testing.T) {
	g := newCatalogGame(t)
	id := crHydra(t, g, 5)
	crRemove(t, g, id, game.CounterPlusOne, 3)
	if g.Stack.Size()+len(g.PendingTriggers) != 3 {
		t.Fatalf("%d stack items after three counters were removed at once, want 3", g.Stack.Size())
	}
	crEndOfTurn(t, g)
	if got := pr8Counters(g, id, game.CounterPlusOne); got != 2+6 {
		t.Fatalf("counters: got %d, want 8 (5, three removed, six back at end step)", got)
	}
}

func TestRemovingAnotherKindOfCounterDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	id := crHydra(t, g, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterCharge, 2); err != nil {
			t.Fatal(err)
		}
	})
	crRemove(t, g, id, game.CounterCharge, 2)
	if g.Stack.Size()+len(g.PendingTriggers) != 0 {
		t.Fatalf("%d stack items after a charge counter was removed, want 0", g.Stack.Size())
	}
	crEndOfTurn(t, g)
	if got := pr8Counters(g, id, game.CounterPlusOne); got != 2 {
		t.Fatalf("counters: got %d, want 2", got)
	}
}

// Placing a counter is not removing one.
func TestPlacingACounterDoesNotTriggerTheRemovalAbility(t *testing.T) {
	g := newCatalogGame(t)
	id := crHydra(t, g, 2)
	putCounters(t, g, id, game.CounterPlusOne, 3)
	if g.Stack.Size()+len(g.PendingTriggers) != 0 {
		t.Fatalf("%d stack items after a placement, want 0", g.Stack.Size())
	}
}

// CR 122.2: counters go away with a permanent that leaves; they are not
// "removed", so nothing triggers, and the Hydra's removal ability is gone
// with it anyway. The witness is a second removal-watcher that stays: a
// Hydra killed with counters on it makes no end-step counters for itself.
func TestACounterCardLeavingTheBattlefieldIsNotARemoval(t *testing.T) {
	g := newCatalogGame(t)
	id := crHydra(t, g, 3)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if findBattlefieldCardForTest(g, id) != nil {
		t.Fatal("the Hydra should be gone")
	}
	// No counter event after the destruction names the card.
	var after bool
	g.ReadSnapshot(func() {
		seenDeath := false
		for _, ev := range g.Events {
			if ev.Kind == game.EventLTB && ev.CardID == id {
				seenDeath = true
			}
			if seenDeath && ev.Kind == game.EventCounterPlaced && ev.Target == id {
				after = true
			}
		}
	})
	if after {
		t.Fatal("leaving the battlefield emitted a counter-removal event")
	}
	if g.Stack.Size()+len(g.PendingTriggers) != 0 {
		t.Fatalf("%d stack items after the Hydra left, want 0", g.Stack.Size())
	}
}

// Prevention is itself a removal, so the Hydra's own clause feeds its
// trigger: two damage takes two counters off and puts four back.
func TestProteanHydraDamagePreventionRemovesThenReturnsAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	id := crHydra(t, g, 5)
	src := pr7Creature(g, opp.ID, "Bear", 2, "G")
	pr6Damage(t, g, src, id, 2)
	if pr6Marked(g, id) != 0 {
		t.Fatalf("damage marked %d, want 0 (prevented)", pr6Marked(g, id))
	}
	if got := pr8Counters(g, id, game.CounterPlusOne); got != 3 {
		t.Fatalf("counters after the hit: got %d, want 3", got)
	}
	crEndOfTurn(t, g)
	if got := pr8Counters(g, id, game.CounterPlusOne); got != 3+4 {
		t.Fatalf("counters at end step: got %d, want 7 (two triggers, two counters each)", got)
	}
}

// Damage past the last counter: every counter comes off, the damage is
// still prevented, and the Hydra dies as a 0/0 before any end step —
// the delayed triggers find no Hydra and do nothing.
func TestProteanHydraRemovedBeyondItsCountersDiesAndGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	id := pr8cPermanent(t, g, g.Seats[0].ID, "Protean Hydra", oracleProteanHydra, 0, 0, 2)
	src := pr7Creature(g, opp.ID, "Bear", 5, "G")
	pr6Damage(t, g, src, id, 5)
	crEndOfTurn(t, g)
	if findBattlefieldCardForTest(g, id) != nil {
		t.Fatal("a 0/0 Hydra with no counters should have died")
	}
}

func TestProteanHydraDeclarationIsFull(t *testing.T) {
	spec, ok := Lookup(oracleProteanHydra)
	if !ok {
		t.Fatal("Protean Hydra is not in the catalog")
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("completeness %v, want full", spec.Completeness)
	}
	if len(spec.Triggered) != 1 || spec.Triggered[0].PerCounterRemoved != game.CounterPlusOne {
		t.Errorf("Protean Hydra's trigger should be PerCounterRemoved +1/+1")
	}
}

// --- Dinosaurs on a Spaceship ---------------------------------------------

func crSuspended(t *testing.T, g *game.Game, counters int) uuid.UUID {
	me := g.Seats[0]
	id := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: id, Name: "Dinosaurs on a Spaceship", OracleID: oracleDinosaursSpaceSh,
		TypeLine: "Creature — Dinosaur", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	// Through the counter primitive, as suspend itself places them, so
	// the log has the placement the removal reading starts from.
	putCounters(t, g, id, game.CounterTime, counters)
	return id
}

func crDinosaurTokens(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dinosaur" && c.IsToken() {
			n++
		}
	}
	return n
}

func TestDinosaursOnASpaceshipOneTimeCounterIsOneDinosaur(t *testing.T) {
	g := newCatalogGame(t)
	id := crSuspended(t, g, 4)
	crRemove(t, g, id, game.CounterTime, 1)
	passPriorityAroundTable(t, g)
	if got := crDinosaurTokens(g); got != 1 {
		t.Fatalf("%d Dinosaur tokens, want 1", got)
	}
}

func TestDinosaursOnASpaceshipThreeTimeCountersAreThreeDinosaurs(t *testing.T) {
	g := newCatalogGame(t)
	id := crSuspended(t, g, 4)
	crRemove(t, g, id, game.CounterTime, 3)
	passPriorityAroundTable(t, g)
	if got := crDinosaurTokens(g); got != 3 {
		t.Fatalf("%d Dinosaur tokens, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dinosaur" && (c.Power != 2 || c.Toughness != 2 || len(c.Colors) != 2) {
			t.Fatalf("token %+v is not a 2/2 red and white Dinosaur", c)
		}
	}
}

// The ability reads "while it's exiled": the same card on the battlefield
// with time counters on it does nothing when one comes off.
func TestDinosaursOnASpaceshipDoesNothingOffExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := apaPush(g, me.ID, me.ID, game.Card{Name: "Dinosaurs on a Spaceship", OracleID: oracleDinosaursSpaceSh, TypeLine: "Creature — Dinosaur", Power: 6, Toughness: 6})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterTime, 2) })
	crRemove(t, g, id, game.CounterTime, 1)
	passPriorityAroundTable(t, g)
	if got := crDinosaurTokens(g); got != 0 {
		t.Fatalf("%d Dinosaur tokens from the battlefield, want 0", got)
	}
}

// A different kind of counter, even in exile, is nothing.
func TestDinosaursOnASpaceshipIgnoresOtherCounters(t *testing.T) {
	g := newCatalogGame(t)
	id := crSuspended(t, g, 4)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterCharge, 2) })
	crRemove(t, g, id, game.CounterCharge, 2)
	passPriorityAroundTable(t, g)
	if got := crDinosaurTokens(g); got != 0 {
		t.Fatalf("%d Dinosaur tokens, want 0", got)
	}
}

func TestDinosaursOnASpaceshipDeclaration(t *testing.T) {
	sa := suspendOf(oracleDinosaursSpaceSh)
	if sa == nil || sa.Counters != 4 || sa.Cost != "{3}{R}{W}" {
		t.Fatalf("suspend declaration %+v, want 4 / {3}{R}{W}", sa)
	}
	spec, ok := Lookup(oracleDinosaursSpaceSh)
	if !ok || spec.Completeness != CompletenessFull {
		t.Fatal("Dinosaurs on a Spaceship should be catalogued and Full")
	}
}

// The lord: other Dinosaurs you control get +1/+1 and vigilance and
// trample; non-Dinosaurs and opponents' do not.
func TestDinosaursOnASpaceshipLordsOtherDinosaurs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := apaPush(g, me.ID, me.ID, game.Card{Name: "Dinosaurs on a Spaceship", OracleID: oracleDinosaursSpaceSh, TypeLine: "Creature — Dinosaur", Power: 6, Toughness: 6})
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "Raptor", TypeLine: "Creature — Dinosaur", Power: 2, Toughness: 2})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Raptor", TypeLine: "Creature — Dinosaur", Power: 2, Toughness: 2})
	if got := effectivePower(t, g, mine); got != 3 {
		t.Errorf("my Dinosaur power %d, want 3", got)
	}
	if !effectiveAbilitiesContain(t, g, mine, "vigilance") || !effectiveAbilitiesContain(t, g, mine, "trample") {
		t.Errorf("my Dinosaur should have vigilance and trample")
	}
	if got := effectivePower(t, g, lord); got != 6 {
		t.Errorf("the lord's own power %d, want 6 (other Dinosaurs)", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("the Bear's power %d, want 2", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("their Dinosaur's power %d, want 2", got)
	}
}
