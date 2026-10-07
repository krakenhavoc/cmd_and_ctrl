package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// return_self_cost_test.go — the enumerator half of #2028, "Return this
// <permanent> to its owner's hand" as a cost (AbilityCost.ReturnSelf).
// The #544 invariant: every move offered is one the engine accepts, and
// none spends the source twice.

func returnSelfRow(cost game.AbilityCost) []game.ActivatedAbilityShape {
	cost.ReturnSelf = true
	return []game.ActivatedAbilityShape{{
		Label:  "Return this to its owner's hand: nothing.",
		Cost:   cost,
		Effect: func(*game.Game, *game.StackItem) error { return nil },
	}}
}

// The plain cost: offered, and the offer is accepted.
func TestReturnSelfIsEnumeratedAndAccepted(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	chains := battlefieldCard(g, active, game.Card{
		Name: "Bouncing Charm", TypeLine: "Enchantment",
		ActivatedAbilities: returnSelfRow(game.AbilityCost{}),
	})
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), chains)
	if len(acts) != 1 {
		t.Fatalf("want exactly one return-this move, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)
}

// The source is never also a sacrifice pick: the sacrifice clause admits
// it, but it is already paying the return (CR 118.3).
func TestReturnSelfSourceIsNotOfferedAsASacrifice(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	src := battlefieldCard(g, active, game.Card{
		Name: "Bouncing Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		ActivatedAbilities: returnSelfRow(game.AbilityCost{SacrificeOther: &game.TargetSpec{
			Mode:  "permanent",
			Label: "a creature you control",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1, Max: 1,
		}}),
	})
	other := battlefieldCard(g, active, game.Card{Name: "Rat", TypeLine: "Creature — Rat", Power: 1, Toughness: 1})
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), src)
	if len(acts) != 1 {
		t.Fatalf("want one move (sacrifice the Rat), got %v", labels(acts))
	}
	var p struct {
		SacrificeIDs []string `json:"sacrifice_ids"`
	}
	if err := json.Unmarshal(acts[0].Params, &p); err != nil {
		t.Fatalf("params: %v", err)
	}
	if len(p.SacrificeIDs) != 1 || p.SacrificeIDs[0] != other.String() {
		t.Fatalf("sacrifice_ids = %v, want only the Rat", p.SacrificeIDs)
	}
	dispatchAll(t, g, active.ID, acts)
}

// A land that returns itself cannot also tap for the mana its own return
// costs: the planner keeps the source out (AbilityAutoTapExclusions). With
// only itself to pay, the ability is not offered; with a second land, it
// is, and the engine accepts it.
func TestReturnSelfSourceIsNotTappedForItsOwnMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	land := battlefieldCard(g, active, game.Card{
		Name: "Bouncing Island", TypeLine: "Basic Land — Island",
		ActivatedAbilities: returnSelfRow(game.AbilityCost{Mana: "{1}"}),
	})
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), land); len(acts) != 0 {
		t.Fatalf("offered with only the returning land to pay: %v", labels(acts))
	}
	mana(g, active, 1)
	acts := activationsOf(legal.EnumerateFor(g, active.ID), land)
	if len(acts) != 1 {
		t.Fatalf("want one move with a second land, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)
}
