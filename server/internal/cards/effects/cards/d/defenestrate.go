package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defenestrate — Instant {2}{B} (Innistrad: Midnight Hunt, #5497):
//
//	"Destroy target creature without flying."
//
// Doom Blade's shape; flying is read through the layered view, so a creature
// granted flying is not a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caf840e8-c1d3-4570-8542-432a765c09ff",
		Name:         "Defenestrate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature without flying", effects.WithoutKeyword("flying")),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.S82MidDestroyTheTarget,
	})
}
