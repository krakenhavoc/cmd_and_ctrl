package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Debtors' Transport — Creature — Thrull {5}{B}, 5/3 (Ravnica Allegiance,
// #5521):
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
		OracleID:     "72f89173-2998-45e6-94da-5b5bb9738a6b",
		Name:         "Debtors' Transport",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Debtors' Transport — afterlife 2: create two 1/1 white and black Spirit creature tokens with flying",
				effects.KnightOfTheLastBreathSpirits(2)),
		},
	})
}
