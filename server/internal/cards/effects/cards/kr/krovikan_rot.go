package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krovikan Rot — Instant {2}{B} (Coldsnap):
//
//	"Destroy target creature with power 2 or less.
//	 Recover {1}{B}{B}"
//
// The power is the target's as it is now, judged on casting and again as the
// spell resolves (CR 608.2b). Recover (CR 702.59a) is S82OldB564Recover.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9a272c5-6ebc-4d29-a47c-a2d570184196",
		Name:         "Krovikan Rot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
		OnResolve:    effects.DestroyTheTargetPermanent,
		Triggered:    []game.TriggeredAbility{effects.S82OldB564Recover("Krovikan Rot", "{1}{B}{B}")},
	})
}
