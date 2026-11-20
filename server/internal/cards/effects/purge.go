package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d06f0f7c-0270-44be-991a-a5ebbddb0900",
		Name:         "Purge",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target artifact creature or black creature", Or(Artifact(), OfColor("B"))),
		OnResolve:    destroyTheTargetPermanentNoRegen,
	})
}
