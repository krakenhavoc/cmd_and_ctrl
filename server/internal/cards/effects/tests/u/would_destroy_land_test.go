package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// would_destroy_land_test.go — #6225: Equinox's granted ability,
// "{T}: Counter target spell if it would destroy a land you control."
// The ability is registered here on a test land, the shape Equinox's
// enchanted land has once its Aura grants it; the card itself is its
// batch's. CR 701.8a: "To destroy a permanent, move it from the
// battlefield to its owner's graveyard."

const (
	equinoxTestOracle = "test-6225-would-destroy-land"
	stoneRainOracle   = "6e880df6-fc00-43d2-a9c8-f575f40b78c6"
	shockOracle       = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
	armageddonOracle  = "c9ed8b01-959a-47d6-891e-0abbdccf6e4f"
	murderOracle      = "938b4e2c-88d9-4637-bc00-e228920c9a78"
)

func registerEquinoxLand(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: equinoxTestOracle,
		Name:     "Equinox Test Land",
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Counter target spell if it would destroy a land you control.",
			Cost:    game.AbilityCost{Tap: true},
			Targets: effects.TargetSpell("target spell"),
			Effect:  effects.CounterTargetSpellIfItWouldDestroyALandYouControl,
		}},
	})
}

// equinoxTable is a game in seat 0's main phase where seat 1 controls a
// Forest and the Equinox land.
type equinoxTable struct {
	g              *game.Game
	me, opp        *game.Player
	myLand, forest uuid.UUID
	guard          uuid.UUID
}

func newEquinoxTable(t *testing.T) equinoxTable {
	t.Helper()
	registerEquinoxLand(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	return equinoxTable{
		g: g, me: me, opp: opp,
		myLand: effectstest.FreeSpellPermanent(g, me.ID, "Mountain", "Basic Land — Mountain"),
		forest: effectstest.FreeSpellPermanent(g, opp.ID, "Forest", "Basic Land — Forest"),
		guard:  effectstest.PushCatalogPermanent(g, opp.ID, "Equinox Test Land", "Land", equinoxTestOracle, false),
	}
}

// respond passes seat 0's priority, has seat 1 activate the ability at
// the spell, and passes until the ability has resolved.
func (e equinoxTable) respond(t *testing.T, spell uuid.UUID) {
	t.Helper()
	if err := e.g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := effectstest.ActivateOn(t, e.g, e.opp.ID, e.guard, 0, spell); err != nil {
		t.Fatalf("activating at the spell: %v", err)
	}
	items := len(e.g.StackMeta)
	for i := 0; i < 8 && len(e.g.StackMeta) >= items; i++ {
		if err := e.g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if len(e.g.StackMeta) >= items {
		t.Fatal("the ability never resolved")
	}
}

func TestEquinoxCountersAStoneRainAtALandYouControl(t *testing.T) {
	e := newEquinoxTable(t)
	rain := effectstest.CastCatalogSpell(t, e.g, "Stone Rain", "Sorcery", stoneRainOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: e.forest}})
	e.respond(t, rain)
	if e.g.Stack.Contains(rain) {
		t.Fatal("Stone Rain at the Forest should have been countered")
	}
	if !e.me.Graveyard.Contains(rain) {
		t.Error("the countered Stone Rain should be in its owner's graveyard")
	}
	if !e.g.Battlefield.Contains(e.forest) {
		t.Error("the Forest should still be on the battlefield")
	}
}

func TestEquinoxLeavesAShockAlone(t *testing.T) {
	e := newEquinoxTable(t)
	shock := effectstest.CastCatalogSpell(t, e.g, "Shock", "Instant", shockOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: e.opp.ID}})
	e.respond(t, shock)
	if !e.g.Stack.Contains(shock) {
		t.Fatal("Shock destroys no land and must not be countered")
	}
}

// "A land YOU control": a Stone Rain at someone else's land is not
// countered by the opponent's Equinox, and the land still dies.
func TestEquinoxLeavesAStoneRainAtSomeoneElsesLand(t *testing.T) {
	e := newEquinoxTable(t)
	rain := effectstest.CastCatalogSpell(t, e.g, "Stone Rain", "Sorcery", stoneRainOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: e.myLand}})
	e.respond(t, rain)
	if !e.g.Stack.Contains(rain) {
		t.Fatal("Stone Rain at seat 0's own land must not be countered by seat 1's Equinox")
	}
	effectstest.PassPriorityAroundTable(t, e.g)
	if e.g.Battlefield.Contains(e.myLand) {
		t.Error("the targeted Mountain should have been destroyed")
	}
}

// An untargeted sweep counts: Armageddon destroys all lands.
func TestEquinoxCountersArmageddon(t *testing.T) {
	e := newEquinoxTable(t)
	geddon := effectstest.CastCatalogSpell(t, e.g, "Armageddon", "Sorcery", armageddonOracle, nil)
	e.respond(t, geddon)
	if e.g.Stack.Contains(geddon) {
		t.Fatal("Armageddon would destroy seat 1's lands and should have been countered")
	}
	if !e.g.Battlefield.Contains(e.myLand) || !e.g.Battlefield.Contains(e.forest) {
		t.Error("no land should have been destroyed")
	}
}

// A land creature is a land: Murder at it would destroy a land.
func TestEquinoxCountersMurderAtALandCreature(t *testing.T) {
	e := newEquinoxTable(t)
	arbor := effectstest.PushCatalogPermanent(e.g, e.opp.ID, "Arbor Test", "Land Creature — Forest Dryad", "", false)
	murder := effectstest.CastCatalogSpell(t, e.g, "Murder", "Instant", murderOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: arbor}})
	e.respond(t, murder)
	if e.g.Stack.Contains(murder) {
		t.Fatal("Murder at a land creature would destroy a land and should have been countered")
	}
	if !e.g.Battlefield.Contains(arbor) {
		t.Error("the land creature should still be on the battlefield")
	}
}

// The question is answered on a copy: asking it changes nothing in the
// live game, not the stack, the board, the log nor the random stream.
func TestWouldDestroyIsAskedOfACopy(t *testing.T) {
	e := newEquinoxTable(t)
	rain := effectstest.CastCatalogSpell(t, e.g, "Stone Rain", "Sorcery", stoneRainOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: e.forest}})
	events, stack := len(e.g.Events), len(e.g.Stack.Cards)
	var hit, miss bool
	e.g.WithWriteLock(func() {
		hit = e.g.SpellWouldDestroyALandOfForEffect(rain, e.opp.ID)
		miss = e.g.SpellWouldDestroyALandOfForEffect(rain, e.me.ID)
	})
	if !hit || miss {
		t.Fatalf("would destroy a land of seat 1 = %v (want true), of seat 0 = %v (want false)", hit, miss)
	}
	if len(e.g.Events) != events || len(e.g.Stack.Cards) != stack || !e.g.Stack.Contains(rain) || !e.g.Battlefield.Contains(e.forest) {
		t.Error("asking what the spell would destroy changed the live game")
	}
}
