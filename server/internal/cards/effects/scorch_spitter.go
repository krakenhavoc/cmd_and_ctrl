package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorch Spitter — Creature — Elemental Lizard {R}, 1/1 (Core Set 2020, #5518):
//
//	"Whenever this creature attacks, it deals 1 damage to the player or
//	 planeswalker it's attacking."
//
// The attacked player or planeswalker is the one the attack was declared
// against, carried on the trigger's event (Hellrider's reading), so the
// damage follows it even if the Spitter leaves combat in response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4cdc2120-7051-45b8-b877-87ffaf396ee7",
		Name:         "Scorch Spitter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Scorch Spitter — 1 damage to the player or planeswalker it's attacking",
				func(g *game.Game, item *game.StackItem) error {
					return DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Target, Amount: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
