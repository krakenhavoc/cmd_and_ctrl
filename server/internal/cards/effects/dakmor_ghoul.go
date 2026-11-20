package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6b25b7d7-ebb1-474c-9898-e027dbf0bbdd",
		Name:         "Dakmor Ghoul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Dakmor Ghoul — target opponent loses 2 life and you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return targetOpponentLosesAndYouGain(g, item, 2)
					}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
