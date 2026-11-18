package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defenestrate — Instant {2}{B} (Innistrad: Midnight Hunt, #5497):
//
//	"Destroy target creature without flying."
//
// Doom Blade's shape; flying is read through the layered view, so a creature
// granted flying is not a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "caf840e8-c1d3-4570-8542-432a765c09ff",
		Name:         "Defenestrate",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature without flying", WithoutKeyword("flying")),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    s82MidDestroyTheTarget,
	})
}
