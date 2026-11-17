package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nettletooth Djinn — Creature — Djinn {3}{G}, 4/4 (Mirage, #5633):
//
//	"At the beginning of your upkeep, this creature deals 1 damage to
//	 you."
//
// Serendib Efreet's upkeep trigger: damage from the Djinn to its
// controller, dealt by the Djinn as it last existed if it has left the
// battlefield in response (CR 608.2h). It is damage, so a prevention
// effect can stop it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1d300b6-f9cf-40a9-8520-d78cd7d813cf",
		Name:         "Nettletooth Djinn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Nettletooth Djinn — this creature deals 1 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
