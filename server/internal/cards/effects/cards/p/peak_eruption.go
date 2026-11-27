package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peak Eruption — Sorcery {2}{R} (Theros, S82 OLD b392, #5561):
//
//	"Destroy target Mountain. Peak Eruption deals 3 damage to that land's
//	 controller."
//
// Destructive Revelry's destroy-then-damage body (the controller is read
// before the destruction) over a Mountain target, judged at announce and again
// at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f34eb8a9-63af-452b-8957-8587bf4fd956",
		Name:         "Peak Eruption",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target Mountain", effects.And(effects.Land(), effects.Subtype("Mountain"))),
		OnResolve:    effects.DestroyTargetThenDamageItsController(3),
	})
}
