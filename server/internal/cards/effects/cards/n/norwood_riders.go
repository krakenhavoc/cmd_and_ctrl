package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Norwood Riders — Creature — Elf {3}{G}, 3/3 (Portal Second Age, S82 older
// sets batch 733, #5625):
//
//	"This creature can't be blocked by more than one creature."
//
// One block-count rule (MaxBlockers, CR 509.1b) on itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1816b87c-73ba-4bfc-a51e-c982381c067d",
		Name:         "Norwood Riders",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
	})
}
