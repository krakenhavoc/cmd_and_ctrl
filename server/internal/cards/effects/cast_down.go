package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cast Down — Instant {1}{B} (Dominaria, #5530):
//
//	"Destroy target nonlegendary creature."
//
// "Destroy target nonlegendary creature." — the shared
// destroyTheTargetPermanent body; the supertype is read at announce and
// again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cdaab6b0-1a2d-4809-8e6b-56013acd8f78",
		Name:         "Cast Down",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target nonlegendary creature", Not(Legendary())),
		OnResolve:    destroyTheTargetPermanent,
	})
}
