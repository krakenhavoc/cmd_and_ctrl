package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triumph of Cruelty — Enchantment {2}{B} (Avacyn Restored, S82 OLD, #5570):
//
//	"At the beginning of your upkeep, target opponent discards a card if you
//	 control the creature with the greatest power or tied for the greatest
//	 power."
//
// Triumph of Ferocity's condition (read as the ability resolves, not an
// intervening "if", so the opponent is targeted either way) and a discard
// the opponent chooses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87e88311-dbd2-42f7-a166-138f1d271686",
		Name:         "Triumph of Cruelty",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Triumph of Cruelty — target opponent discards a card if you control the creature with the greatest power or tied for the greatest power",
					func(g *game.Game, item *game.StackItem) error {
						if !effects.ControlsGreatestPower(g, item.Controller) {
							return nil
						}
						return effects.TargetOpponentDiscards(1)(g, item)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent()),
			),
		},
	})
}
