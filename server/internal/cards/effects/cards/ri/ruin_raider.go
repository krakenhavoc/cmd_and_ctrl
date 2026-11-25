package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruin Raider — Creature — Orc Pirate {2}{B}, 3/2 (Ixalan):
//
//	"Raid — At the beginning of your end step, if you attacked this turn,
//	 reveal the top card of your library and put that card into your hand. You
//	 lose life equal to the card's mana value."
//
// Dark Tutelage's flip (Dark Confidant's): the reveal is not a draw, the mana
// value is read before the move, a land costs nothing and an empty library
// does nothing. The raid is Perforating Artist's end-step intervening "if".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6a837d0-d4f9-45ed-bbd7-5e37a9ca3b56",
		Name:         "Ruin Raider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB223RaidAtYourEndStep("Ruin Raider — raid: reveal the top card of your library and put it into your hand; lose life equal to its mana value",
				effects.DarkConfidantFlip),
		},
	})
}
