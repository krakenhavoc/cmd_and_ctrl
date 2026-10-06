package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ouroboroidOracle = "50d6fd91-23d3-4d32-804f-6233e4386904"

// TestOuroboroidPutsItsOwnPowerInCountersOnEachCreatureAtCombat — a
// 1/3 base Ouroboroid puts one counter on each creature the
// controller controls (itself included) at the beginning of combat.
func TestOuroboroidPutsItsOwnPowerInCountersOnEachCreatureAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ouro := pushCatalogPermanent(g, me.ID, "Ouroboroid", "Creature — Plant Wurm", ouroboroidOracle, false)
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)

	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	var ouroC, mineC, theirsC int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			switch c.InstanceID {
			case ouro:
				ouroC = c.Counters[game.CounterPlusOne]
			case mine:
				mineC = c.Counters[game.CounterPlusOne]
			case theirs:
				theirsC = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if ouroC != 1 {
		t.Errorf("Ouroboroid's own counters = %d, want 1 (its base power is 1)", ouroC)
	}
	if mineC != 1 {
		t.Errorf("controller's other creature counters = %d, want 1", mineC)
	}
	if theirsC != 0 {
		t.Errorf("opponent's creature counters = %d, want 0 (not \"you control\")", theirsC)
	}
}

// ouroboroidCreature pushes a vanilla creature with printed P/T. The
// #2401 board's cards stand in as these: their own text plays no part
// in the counters.
func ouroboroidCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// ouroboroidSetCounters sets a battlefield card's +1/+1 counters
// directly, standing in for counters it got earlier in the game.
func ouroboroidSetCounters(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: n}
			}
		}
	})
}

// ouroboroidToCombat walks to seat 0's next beginning of combat (it
// always leaves the current step first) without resolving anything.
func ouroboroidToCombat(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepBeginCombat {
			return
		}
	}
	t.Fatal("never reached seat 0's beginning of combat")
}

// TestOuroboroidCountsTheCountersOnItInX is #2401's board. Ouroboroid
// came down before three of its controller's combats; Circuit Mender
// arrived after the first and Memnite after the second. "X is this
// creature's power" includes the counters on it (CR 122.1a), so X goes
// 1, 2, 4 and Ouroboroid ends with 7 counters. The bug read the power
// without its counters, X stayed 1, and the table showed exactly the
// report's screenshot: 3 counters on Ouroboroid, Ornithopter of
// Paradise and Elvish Mystic, 4 on Marwyn (one from an Elf entering),
// 2 on Circuit Mender and 1 on Memnite.
func TestOuroboroidCountsTheCountersOnItInX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ouro := pushCatalogPermanent(g, me.ID, "Ouroboroid", "Creature — Plant Wurm", ouroboroidOracle, false)
	thopter := ouroboroidCreature(g, me.ID, "Ornithopter of Paradise", 0, 2)
	mystic := ouroboroidCreature(g, me.ID, "Elvish Mystic", 1, 1)
	marwyn := ouroboroidCreature(g, me.ID, "Marwyn, the Nurturer", 1, 1)
	ouroboroidSetCounters(g, marwyn, 1)

	ouroboroidToCombat(t, g) // X = 1
	passPriorityAroundTable(t, g)
	mender := ouroboroidCreature(g, me.ID, "Circuit Mender", 2, 3)
	ouroboroidToCombat(t, g) // X = 2
	passPriorityAroundTable(t, g)
	memnite := ouroboroidCreature(g, me.ID, "Memnite", 1, 1)
	ouroboroidToCombat(t, g) // X = 4
	passPriorityAroundTable(t, g)

	for _, want := range []struct {
		name string
		id   uuid.UUID
		n    int
	}{
		{"Ouroboroid", ouro, 7},
		{"Ornithopter of Paradise", thopter, 7},
		{"Elvish Mystic", mystic, 7},
		{"Marwyn, the Nurturer", marwyn, 8},
		{"Circuit Mender", mender, 6},
		{"Memnite", memnite, 4},
	} {
		if got := counterCount(g, want.id, game.CounterPlusOne); got != want.n {
			t.Errorf("%s: +1/+1 counters = %d, want %d", want.name, got, want.n)
		}
	}
}

// TestOuroboroidUsesItsLastKnownPowerWhenItHasLeft: Ouroboroid with two
// counters (power 3) is sacrificed with its trigger on the stack. The
// trigger still resolves, and X is its power as it last existed on the
// battlefield (CR 608.2h), counters included.
func TestOuroboroidUsesItsLastKnownPowerWhenItHasLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ouro := pushCatalogPermanent(g, me.ID, "Ouroboroid", "Creature — Plant Wurm", ouroboroidOracle, false)
	bear := seedCreature(g, "My Bear", me.ID)
	ouroboroidSetCounters(g, ouro, 2)

	ouroboroidToCombat(t, g)
	if stackFullyEmpty(g) {
		t.Fatal("Ouroboroid's trigger is not on the stack at the beginning of combat")
	}
	var sacErr error
	g.WithWriteLock(func() { sacErr = g.SacrificePermanentForEffect(ouro) })
	if sacErr != nil {
		t.Fatalf("sacrifice Ouroboroid: %v", sacErr)
	}
	passPriorityAroundTable(t, g)

	if got := counterCount(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("bear's +1/+1 counters = %d, want 3 (Ouroboroid's last-known power, counters included)", got)
	}
}

// TestOuroboroidReadsXOnceAndHardenedScalesAddsToEachPlacement: with
// one counter, Ouroboroid's power is 2. X is read once, before any
// counter goes on (CR 608.2h), so Ouroboroid's own new counters do not
// raise what the others get. Hardened Scales adds one to each
// creature's placement, so every creature gets 3.
func TestOuroboroidReadsXOnceAndHardenedScalesAddsToEachPlacement(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ouro := pushCatalogPermanent(g, me.ID, "Ouroboroid", "Creature — Plant Wurm", ouroboroidOracle, false)
	bear := seedCreature(g, "My Bear", me.ID)
	ouroboroidSetCounters(g, ouro, 1)

	ouroboroidToCombat(t, g)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, ouro, game.CounterPlusOne); got != 4 {
		t.Errorf("Ouroboroid's +1/+1 counters = %d, want 4 (1, then X = 2 plus Hardened Scales' 1)", got)
	}
	if got := counterCount(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("bear's +1/+1 counters = %d, want 3 (X = 2, read before Ouroboroid's own counters, plus Hardened Scales' 1)", got)
	}
}
