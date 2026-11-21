package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Lackey — Creature — Goblin {R}, 1/1 (Mystery Booster 2):
//
//	"Whenever this creature deals damage to a player, you may put a Goblin
//	 permanent card from your hand onto the battlefield."
//
// Any damage, not only combat damage. The put is a "you may" the controller
// answers as the trigger resolves, from the Goblin permanent cards in hand
// (CR 110.4; a Goblin instant or sorcery isn't one); it is a put, not a
// cast, so a Goblin enters whatever its cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "213fbbdd-f1bc-4bba-9124-7b26e9a706c8",
		Name:         "Goblin Lackey",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtDamageToAPlayer,
				"Goblin Lackey — you may put a Goblin permanent card from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return PutFromHandOntoBattlefield{
						Match:    Subtype("Goblin"),
						Optional: true,
						Label:    "Goblin Lackey — you may put a Goblin permanent card from your hand onto the battlefield",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
