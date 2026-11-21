package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Purge — Instant {1}{W} (Darksteel, #5605):
//
//	"Destroy target artifact creature or black creature. It can't be
//	 regenerated."
//
// Terror's no-regeneration destroy with the opposite target: a creature
// that is an artifact or is black, judged at announce and again at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d06f0f7c-0270-44be-991a-a5ebbddb0900",
		Name:         "Purge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target artifact creature or black creature", effects.Or(effects.Artifact(), effects.OfColor("B"))),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}
