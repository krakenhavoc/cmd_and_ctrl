package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mikey_and_don_party_planners_test.go — Mikey & Don, Party Planners
// (#4679): lands and Mutant, Ninja or Turtle spells from the top of your
// library, and a creature spell cast that way enters with an additional
// +1/+1 counter (CastPermission.EntersWithCounters).

const mikeyAndDonOracle = "529178e8-425e-4d59-948c-ca7d0663f6ff"

// mikeyTable is the active seat in its main phase with Mikey & Don on
// the battlefield.
func mikeyTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushTopSource(t, g, me.ID, "Mikey & Don, Party Planners", "Legendary Creature — Mutant Ninja Turtle", mikeyAndDonOracle)
	return g, me
}

// mikeyTopTurtle puts a free Turtle creature card on top of `me`'s library.
func mikeyTopTurtle(me *game.Player) uuid.UUID {
	return seedTopOfLibrary(me, game.Card{Name: "Top Turtle", TypeLine: "Creature — Turtle", ManaCost: "{0}", Power: 1, Toughness: 1})
}

// mikeyMarked is the +1/+1 counters the spell's stack item promises, or
// -1 when it is not on the stack.
func mikeyMarked(g *game.Game, spell uuid.UUID) int {
	n := -1
	g.ReadSnapshot(func() {
		item := g.StackMeta[spell]
		if item == nil {
			return
		}
		n = 0
		for _, m := range item.PromisedCounters {
			if m.Kind == game.CounterPlusOne {
				n += m.Count
			}
		}
	})
	return n
}

func mikeyCastFromTop(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast off the top: %v", err)
	}
}

// A Turtle creature cast off the top is marked as it is cast and enters
// with the counter.
func TestMikeyAndDonTurtleCastFromTheTopEntersWithACounter(t *testing.T) {
	g, me := mikeyTable(t)
	turtle := mikeyTopTurtle(me)
	mikeyCastFromTop(t, g, me, turtle)
	if got := mikeyMarked(g, turtle); got != 1 {
		t.Fatalf("the creature spell is marked with %d counters, want 1", got)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(turtle) {
		t.Fatal("the Turtle did not resolve")
	}
	if got := counterCount(g, turtle, game.CounterPlusOne); got != 1 {
		t.Errorf("the Turtle entered with %d +1/+1 counters, want 1", got)
	}
}

// Mutant, Ninja or Turtle: each one is opened, anything else is not, and
// a land on top is playable.
func TestMikeyAndDonOpensMutantsNinjasTurtlesAndLands(t *testing.T) {
	g, me := mikeyTable(t)
	for _, tl := range []string{"Creature — Mutant", "Creature — Human Ninja", "Creature — Turtle"} {
		id := seedTopOfLibrary(me, game.Card{Name: "Top " + tl, TypeLine: tl, ManaCost: "{0}"})
		if !grantedPermissionOn(g, me.ID, id, game.ZoneLibrary).Granted() {
			t.Errorf("%q on top was not opened", tl)
		}
	}
	bear := seedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if grantedPermissionOn(g, me.ID, bear, game.ZoneLibrary).Granted() {
		t.Error("a Bear on top was opened")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Error("a Bear was cast off the top")
	}
	land := seedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play a land off the top: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land played off the top is not on the battlefield")
	}
}

// The rider is the permission's: the same Turtle cast from hand enters
// with none.
func TestMikeyAndDonTurtleCastFromHandGetsNoCounter(t *testing.T) {
	g, me := mikeyTable(t)
	turtle := castFromHandForTest(t, g, me, "Hand Turtle", "Creature — Turtle", "{0}", "", game.CastSpellParams{})
	if got := mikeyMarked(g, turtle); got != 0 {
		t.Fatalf("a Turtle cast from hand is marked with %d counters, want 0", got)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, turtle, game.CounterPlusOne); got != 0 {
		t.Errorf("a Turtle cast from hand entered with %d counters, want 0", got)
	}
}

// "If you cast a creature spell this way": a Turtle instant cast off the
// top is allowed and marked with nothing.
func TestMikeyAndDonNoncreatureTurtleSpellIsNotMarked(t *testing.T) {
	g, me := mikeyTable(t)
	trick := seedTopOfLibrary(me, game.Card{Name: "Turtle Trick", TypeLine: "Kindred Instant — Turtle", ManaCost: "{0}"})
	mikeyCastFromTop(t, g, me, trick)
	if got := mikeyMarked(g, trick); got != 0 {
		t.Errorf("a noncreature spell is marked with %d counters, want 0", got)
	}
}

// The creature spell countered: the mark goes with it, and the next
// Turtle cast from hand gets nothing left over.
func TestMikeyAndDonCounteredSpellLeavesNothingOver(t *testing.T) {
	g, me := mikeyTable(t)
	turtle := mikeyTopTurtle(me)
	mikeyCastFromTop(t, g, me, turtle)
	if err := g.CounterTargetForEffect(turtle); err != nil {
		t.Fatalf("counter the Turtle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(turtle) || g.Battlefield.Contains(turtle) {
		t.Fatal("the countered Turtle should be in its owner's graveyard")
	}
	next := castFromHandForTest(t, g, me, "Next Turtle", "Creature — Turtle", "{0}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, next, game.CounterPlusOne); got != 0 {
		t.Errorf("the next Turtle entered with %d counters, want 0", got)
	}
}

// The counter is an entry counter, so Doubling Season doubles it.
func TestMikeyAndDonCounterIsDoubledByDoublingSeason(t *testing.T) {
	g, me := mikeyTable(t)
	pushCatalogPermanent(g, me.ID, "Doubling Season", "Enchantment", doublingSeasonOracle, false)
	turtle := mikeyTopTurtle(me)
	mikeyCastFromTop(t, g, me, turtle)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, turtle, game.CounterPlusOne); got != 2 {
		t.Errorf("under Doubling Season the Turtle has %d counters, want 2", got)
	}
}

// A restore with the permission open: the restored Mikey & Don still
// opens the top and still gives the counter.
func TestMikeyAndDonPermissionSurvivesARestore(t *testing.T) {
	g, me := mikeyTable(t)
	turtle := mikeyTopTurtle(me)
	restored := restoreThroughJSON(t, g)
	rme := restored.Seats[restored.Turn.ActiveSeat]
	if rme.ID != me.ID {
		t.Fatal("the restored active seat is another player")
	}
	mikeyCastFromTop(t, restored, rme, turtle)
	passPriorityAroundTable(t, restored)
	if got := counterCount(restored, turtle, game.CounterPlusOne); got != 1 {
		t.Errorf("after a restore with the permission open: %d counters, want 1", got)
	}
}

// A restore with the marked spell on the stack: the mark is plain data on
// its item and survives.
func TestMikeyAndDonMarkSurvivesARestore(t *testing.T) {
	g, me := mikeyTable(t)
	turtle := mikeyTopTurtle(me)
	mikeyCastFromTop(t, g, me, turtle)
	restored := restoreThroughJSON(t, g)
	if got := mikeyMarked(restored, turtle); got != 1 {
		t.Fatalf("the restored spell's mark = %d, want 1", got)
	}
	passPriorityAroundTable(t, restored)
	if got := counterCount(restored, turtle, game.CounterPlusOne); got != 1 {
		t.Errorf("after a restore with the marked spell waiting: %d counters, want 1", got)
	}
}
