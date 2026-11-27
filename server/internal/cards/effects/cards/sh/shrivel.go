package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shrivel — Sorcery {1}{B} (Magic 2014, S82 OLD batch 401, #5563):
//
//	"All creatures get -1/-1 until end of turn."
//
// Biting Rain's sweep at -1/-1: one layer-7 modification over every creature on
// the battlefield as it resolves, until cleanup (CR 611.2c, CR 514.2); a
// creature that arrives later is not affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1be478ee-87f3-4265-94d7-08521703c888",
		Name:         "Shrivel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
		OnResolve:    effects.ResolveSteps(effects.BoostUntilEOT{Match: effects.Creature(), Power: -1, Toughness: -1, Label: "Shrivel — -1/-1 until end of turn"}),
	})
}
