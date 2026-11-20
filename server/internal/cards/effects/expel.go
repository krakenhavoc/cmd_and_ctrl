package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expel — Instant {2}{W} (Strixhaven: School of Mages, #5502):
//
//	"Exile target tapped creature."
//
// Assassinate's "target tapped creature" clause with an exile where the
// destroy was. A creature that untaps in response is no longer a legal
// target and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58776b25-794e-4a87-bc71-a77c025b9b03",
		Name:         "Expel",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
