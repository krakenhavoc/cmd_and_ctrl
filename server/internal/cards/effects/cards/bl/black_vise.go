package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Vise — Artifact {1} (Mystery Booster 2):
//
//	"As this artifact enters, choose an opponent.
//	 At the beginning of the chosen player's upkeep, this artifact deals X
//	 damage to that player, where X is the number of cards in their hand
//	 minus 4."
//
// The choice is made as it enters (CR 614.12), so it cannot be answered, and
// is stored on the permanent. The trigger fires only in the chosen player's
// upkeep and counts the hand as it resolves (CR 608.2h); zero or less deals
// nothing. The Rack's shape with Iron Maiden's X.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de7839fb-7040-48ab-a6d4-d1952972943d",
		Name:         "Black Vise",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChoosePlayerAsEnters("Black Vise", effects.Opponents),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.TheChosenPlayersTurn,
				"Black Vise — deals damage to the chosen player equal to their hand size minus 4",
				effects.DamageUpkeepPlayerByHand(effects.HandMinusFour)),
		},
	})
}
