package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Excoriate — Sorcery {3}{W} (Born of the Gods,
// #5559):
//
//	"Exile target tapped creature."
//
// Swallow Whole's exile without its additional cost: the tapped clause
// is on the target (tappedPermanent), checked again as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9aedd96-1aee-480d-87fe-3ed94bd64599",
		Name:         "Excoriate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
