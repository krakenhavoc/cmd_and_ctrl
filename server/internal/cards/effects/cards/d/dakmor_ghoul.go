package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dakmor Ghoul — Creature — Zombie {2}{B}{B}, 2/2 (Starter 1999, #5620):
//
//	"When this creature enters, target opponent loses 2 life and you
//	 gain 2 life."
//
// Bloodborn Scoundrels' enters trigger (targetOpponentLosesAndYouGain):
// the opponent is a Targeting pick, and nothing is lost or gained if
// that opponent has left by the time the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b25b7d7-ebb1-474c-9898-e027dbf0bbdd",
		Name:         "Dakmor Ghoul",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Dakmor Ghoul — target opponent loses 2 life and you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return effects.TargetOpponentLosesAndYouGain(g, item, 2)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
