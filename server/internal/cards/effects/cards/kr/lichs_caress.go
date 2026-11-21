package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lich's Caress — Sorcery {3}{B}{B} (Core Set 2019, #5527):
//
//	"Destroy target creature. You gain 3 life."
//
// The destruction, then the life, in printed order (CR 608.2c),
// through the shared destroyTargetThenGainLife.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01b3af6d-ab7d-4137-a94a-1cf2cbfd9d44",
		Name:         "Lich's Caress",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTargetThenGainLife(3),
	})
}
