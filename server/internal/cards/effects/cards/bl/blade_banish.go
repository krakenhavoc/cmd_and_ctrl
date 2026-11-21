package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blade Banish — Instant {3}{W} (Ikoria: Lair of Behemoths, #5512):
//
//	"Exile target creature with power 4 or greater."
//
// Repel the Vile's first mode as a whole spell: the power is read
// through the layers as it is announced and again as it resolves (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8838e16-96ca-4f73-ad61-7e98a76c4baf",
		Name:         "Blade Banish",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
		OnResolve:    effects.ResolveSteps(effects.ExileChosenTarget{}),
	})
}
