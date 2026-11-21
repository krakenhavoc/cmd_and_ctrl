package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Anarchomancer — Creature — Goblin Shaman {R}{G}, 2/2:
//
//	"Each spell you cast that's red or green costs {1} less to cast."
//
// "Costs" here is CR 601.2f cost reduction over generic mana, same
// vocabulary as every other Medallion-shaped card; the only thing new
// is the two-colour predicate, which cost_modifier.go doesn't have a
// combinator for (And/Or are CardPredicate, not CostPredicate), so
// it's one small closure local to this card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9319f13-1f2b-429c-9d61-e58a3cbec86a",
		Name:         "Goblin Anarchomancer",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Each spell you cast that's red or green costs {1} less to cast.",
				effects.YourSpell(), redOrGreenSpell()),
		},
	})
}

// redOrGreenSpell passes on a spell that is red, green, or both
// (a Gruul spell is both and is still discounted once, since
// CostsLess applies its amount once per matching cast).
func redOrGreenSpell() effects.CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.HasColor(q.Source.TextColor("R")) || q.Card.HasColor(q.Source.TextColor("G"))
	}
}
