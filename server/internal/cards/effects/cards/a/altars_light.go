package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Altar's Light — Instant {2}{W}{W} (Mirrodin, #5606):
//
//	"Exile target artifact or enchantment."
//
// Exorcise's body on an artifact-or-enchantment clause. A target that
// is no longer legal as the spell resolves leaves it with nothing to do
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa9b6be2-b88c-4302-b7e2-faf25a60bcb9",
		Name:         "Altar's Light",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
