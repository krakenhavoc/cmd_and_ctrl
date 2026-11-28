package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Highway Robber — Creature — Human Mercenary {2}{B}{B}, 2/2 (Mercadian
// Masques, #5619):
//
//	"When this creature enters, target opponent loses 2 life and you gain
//	 2 life."
//
// Bloodborn Scoundrels' enters trigger: the opponent is a Targeting pick
// and nothing is lost or gained if that opponent has left by the time the
// trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5de0769e-5141-42d5-aaa6-549cfc154de8",
		Name:         "Highway Robber",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Highway Robber — target opponent loses 2 life and you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return effects.TargetOpponentLosesAndYouGain(g, item, 2)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
