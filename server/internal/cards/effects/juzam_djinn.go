package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Juzám Djinn — Creature — Djinn {2}{B}{B}, 5/5 (Arabian Nights, #5641):
//
//	"At the beginning of your upkeep, this creature deals 1 damage to
//	 you."
//
// Serendib Efreet's upkeep trigger without the flying. The damage is
// dealt by the Djinn even when it has left the battlefield in response,
// using the source as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e81596c-9225-43d1-bd35-798212144f2c",
		Name:         "Juzám Djinn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Juzám Djinn — this creature deals 1 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
