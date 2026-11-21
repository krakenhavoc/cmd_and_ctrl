package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triumphant Surge — Instant {3}{W} (Theros Beyond Death, #5514):
//
//	"Destroy target creature with power 4 or greater. You gain 3
//	 life."
//
// Sanctify's destroy-then-gain-life with Bring Down's power clause. A
// target that is no longer legal makes the spell do nothing, the life
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff281103-7a9b-4eb1-b81e-b8cd137b6e98",
		Name:         "Triumphant Surge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
		OnResolve:    effects.DestroyTargetThenGainLife(3),
	})
}
