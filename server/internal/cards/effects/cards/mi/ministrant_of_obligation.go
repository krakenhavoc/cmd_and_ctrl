package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ministrant of Obligation — Creature — Human Cleric {2}{W}, 2/1 (Ravnica
// Allegiance, #5521):
//
//	"Afterlife 2 (When this creature dies, create two 1/1 white and black
//	 Spirit creature tokens with flying.)"
//
// Afterlife is the dies trigger it reminds (CR 702.135a), Seraph of the
// Scales's two 1/1 white and black Spirits with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0cee5539-d10b-4de2-aee8-e02f0e24ac26",
		Name:         "Ministrant of Obligation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Ministrant of Obligation — afterlife 2: create two 1/1 white and black Spirit creature tokens with flying",
				effects.KnightOfTheLastBreathSpirits(2)),
		},
	})
}
