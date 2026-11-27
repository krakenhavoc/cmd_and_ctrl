package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ironhoof Ox — Creature — Ox {3}{G}{G}, 4/4 (Portal Second Age, S82 older
// sets batch 733, #5625):
//
//	"This creature can't be blocked by more than one creature."
//
// One block-count rule (MaxBlockers, CR 509.1b) on itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1871af55-f7cd-4b8e-affb-5d010df811b1",
		Name:         "Ironhoof Ox",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
	})
}
