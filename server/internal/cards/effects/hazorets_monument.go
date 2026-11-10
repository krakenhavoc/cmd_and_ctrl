package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hazoret's Monument — Legendary Artifact {3}:
//
//	"Red creature spells you cast cost {1} less to cast.
//	 Whenever you cast a creature spell, you may discard a card. If you do,
//	 draw a card."
//
// Oketra's Monument's reduction (red) and cast trigger (any creature spell
// you cast, not only red ones). The loot is Rydia's "you may discard a
// card; if you do, draw a card", asked as the trigger resolves, before the
// creature spell does, and unaffected by that spell being countered.
//
// No simplification.
func init() {
	const label = "Hazoret's Monument — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "f02a4cbf-f0bb-4bc7-99c2-6afa1b41ebb8",
		Name:         "Hazoret's Monument",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Red creature spells you cast cost {1} less to cast.",
				YourSpell(), CreatureSpell(), ColoredSpell("R")),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCast(Creature()), label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
