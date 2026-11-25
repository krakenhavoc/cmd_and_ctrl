package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clockwork Servant — Artifact Creature — Gnome {3}, 2/3 (Throne of Eldraine,
// #5516):
//
//	"Adamant — When this creature enters, if at least three mana of the same
//	 color was spent to cast it, draw a card."
//
// An enters trigger with an intervening "if" (CR 603.4) over the payment of
// the spell that became this permanent (CR 400.7d): checked as the creature
// enters and again as the trigger resolves, so a Servant that has since left
// draws nothing. Colourless is not a colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "67dae813-907e-456d-9e91-2d7c85f1d574",
		Name:           "Clockwork Servant",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		Purpose:        game.Purpose{Draws: 1},
		Triggered:      []game.TriggeredAbility{effects.OldB140ClockworkServantSpentTrigger()},
	})
}
