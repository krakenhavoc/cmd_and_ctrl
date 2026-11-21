package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reprisal — Instant {1}{W} (Alliances, #5634):
//
//	"Destroy target creature with power 4 or greater. It can't be
//	 regenerated."
//
// The power is read when Reprisal is cast and again as it resolves
// (CR 608.2b): a creature shrunk below 4 in response is no longer a
// legal target, and Reprisal does nothing. The destruction (CR 701.8a) is
// not replaced by a regeneration shield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fae90808-bafa-44d8-972a-17ded87576ee",
		Name:         "Reprisal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}
