package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruination Rioter — Creature — Human Berserker {R}{G}, 2/2 (Modern Horizons, #5519):
//
//	"When this creature dies, you may have it deal damage to any target
//	 equal to the number of land cards in your graveyard."
//
// Goblin Arsonist's optional dies trigger with the amount read at
// resolution: the land cards in your graveyard then. The Rioter is the
// source, by its last-known information.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b3cec74-0d32-4762-b3e0-3f4e5eb6f179",
		Name:         "Ruination Rioter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisDies("Ruination Rioter — damage to any target equal to the land cards in your graveyard",
					sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
						return b17LandCardsInGraveyard(ctx.Game, ctx.Controller())
					})),
				TargetAny()), "Ruination Rioter — have it deal damage to any target?"),
		},
	})
}
