package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_permanent_cost_test.go — the enumerator half of #1600's
// exile-a-permanent cost. #544's invariant: one move per creature the
// activator could exile, never an opponent's, none at all when nothing
// can pay — and every move offered is one the dispatcher accepts, on
// both owners of the component.

func exileACreatureCost() *game.ExilePermanentsCost {
	return &game.ExilePermanentsCost{Count: 1, CardType: "creature", Label: "a creature you control"}
}

// soulStoneLike is The Soul Stone's harness shape without the catalog:
// "{1}, {T}, Exile a creature you control: <nothing>".
func soulStoneLike() game.Card {
	return game.Card{
		Name:     "Soul Stone Stand-in",
		TypeLine: "Legendary Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "{1}, {T}, Exile a creature you control: Harness.",
			Cost:   game.AbilityCost{Mana: "{1}", Tap: true, ExilePermanents: exileACreatureCost()},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

// foodChainLike is Food Chain's shape with a fixed output.
func foodChainLike() game.Card {
	return game.Card{
		Name:     "Food Chain Stand-in",
		TypeLine: "Enchantment",
		ManaAbilities: []game.ManaAbilityShape{{
			ExilePermanents: exileACreatureCost(),
			Produced:        "{G}{G}",
			Label:           "Exile a creature you control: Add {G}{G}",
		}},
	}
}

func exiledIDsOf(t *testing.T, m legal.Move) []string {
	t.Helper()
	var p struct {
		ExilePermanentIDs []string `json:"exile_permanent_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return p.ExilePermanentIDs
}

func TestExilePermanentCostIsEnumeratedOncePerCreatureYouControl(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	battlefieldCard(g, active, basic("Forest", "Forest"))
	stone := battlefieldCard(g, active, soulStoneLike())
	bear := battlefieldCard(g, active, creature("Bear", "{1}{G}", 2, 2))
	ogre := battlefieldCard(g, active, creature("Ogre", "{2}{R}", 3, 3))
	theirs := battlefieldCard(g, opp, creature("Their Bear", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, stone)
	if len(acts) != 2 {
		t.Fatalf("want one activation per creature I control, got %v", labels(acts))
	}
	seen := map[string]bool{}
	for _, m := range acts {
		ids := exiledIDsOf(t, m)
		if len(ids) != 1 {
			t.Fatalf("move %q names %v, want exactly one creature", m.Label, ids)
		}
		if ids[0] == theirs.String() {
			t.Errorf("move %q exiles an opponent's creature", m.Label)
		}
		seen[ids[0]] = true
	}
	if !seen[bear.String()] || !seen[ogre.String()] {
		t.Errorf("offered %v, want both of my creatures", seen)
	}
	dispatchAll(t, g, active.ID, acts)
}

func TestExilePermanentCostIsNotEnumeratedWithNothingToExile(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	battlefieldCard(g, active, basic("Forest", "Forest"))
	stone := battlefieldCard(g, active, soulStoneLike())
	chain := battlefieldCard(g, active, foodChainLike())
	battlefieldCard(g, opp, creature("Their Bear", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, stone); len(acts) != 0 {
		t.Errorf("activation offered with no creature of mine to exile: %v", labels(acts))
	}
	if mana := movesOfKindFor(moves, legal.KindMana, chain); len(mana) != 0 {
		t.Errorf("mana ability offered with no creature of mine to exile: %v", labels(mana))
	}
}

func TestExilePermanentManaAbilityIsEnumeratedAndAccepted(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	chain := battlefieldCard(g, active, foodChainLike())
	bear := battlefieldCard(g, active, creature("Bear", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	mana := movesOfKindFor(legal.EnumerateFor(g, active.ID), legal.KindMana, chain)
	if len(mana) != 1 {
		t.Fatalf("want one mana move, got %v", labels(mana))
	}
	if ids := exiledIDsOf(t, mana[0]); len(ids) != 1 || ids[0] != bear.String() {
		t.Errorf("mana move exiles %v, want the bear", ids)
	}
	dispatchAll(t, g, active.ID, mana)
}
