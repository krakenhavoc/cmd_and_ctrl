package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// False Prophet — Creature — Human Cleric {2}{W}{W}, 2/2 (Commander 2011, S82
// older sets batch 453, #5574):
//
//	"When this creature dies, exile all creatures."
//
// Exiles every creature on the battlefield, the Prophet's controller's
// included; the Prophet is already in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87c44c23-f209-494a-ab66-c9f5cf7bb4ea",
		Name:         "False Prophet",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("False Prophet — exile all creatures",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.ExileAllMatching{Match: effects.Creature()}.Apply(ctx)
				}))),
		},
	})
}
