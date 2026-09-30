package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const walkingBallistaOracle = "4b515bb0-f275-4400-8032-3173b799ab40"

// TestWalkingBallistaEntersWithXCountersAndPingsWithARemoved pins the
// X entry, the {4} counter-add ability, and the remove-a-counter
// damage ability.
func TestWalkingBallistaEntersWithXCountersAndPingsWithARemoved(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	id := castXSpell(t, g, "Walking Ballista", "Artifact Creature — Construct", walkingBallistaOracle, "{X}{X}", 2, nil)
	passPriorityAroundTable(t, g)

	var counters int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				counters = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if counters != 2 {
		t.Fatalf("counters after ETB with X=2: %d, want 2", counters)
	}

	// {4}: put a +1/+1 counter on this creature.
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate {4}: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				counters = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if counters != 3 {
		t.Fatalf("counters after {4} ability: %d, want 3", counters)
	}

	// Remove a +1/+1 counter: deal 1 damage to any target.
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate remove-counter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d -> %d, want -1", before, opp.Life)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				counters = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if counters != 2 {
		t.Errorf("counters after removing one: %d, want 2", counters)
	}
}

// TestWalkingBallistaCantPingWithNoCountersLeft — removing the last
// counter drops Ballista to 0/0 (CR 704.5f); a second ping refuses
// because the creature is no longer on the battlefield to activate.
func TestWalkingBallistaCantPingWithNoCountersLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	id := castXSpell(t, g, "Walking Ballista", "Artifact Creature — Construct", walkingBallistaOracle, "{X}{X}", 1, nil)
	passPriorityAroundTable(t, g)

	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("first ping: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Fatalf("Walking Ballista at 0 counters should be a 0/0 and die to state-based actions")
	}

	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Error("a second ping after Ballista died should be refused")
	}
}
