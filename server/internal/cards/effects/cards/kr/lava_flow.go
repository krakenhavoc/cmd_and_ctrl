package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lava Flow — Sorcery {3}{R}{R} (Portal, S82 OLD batch 784, #5630):
//
//	"Destroy target creature or land."
//
// Wrecking Ball's target and body: any creature or land, destroyed as the spell
// resolves; regeneration and indestructible work as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91c0a76e-3992-437f-b85a-97b0b4adbb84",
		Name:         "Lava Flow",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or land", effects.Or(effects.Creature(), effects.Land())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
