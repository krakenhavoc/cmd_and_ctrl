package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plow Under — Sorcery {3}{G}{G} (Eighth Edition, S82 older sets batch 618, #5607):
//
//	"Put two target lands on top of their owners' libraries."
//
// Anchor to the Aether's tuck for each of two targets. A target that has
// become illegal is skipped (CR 608.2b); the other still goes. Each land goes
// to the top of its OWNER's library, in target order, so the second is on top
// when both share an owner.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23b36aef-961f-4a78-9652-8594ad73102e",
		Name:         "Plow Under",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("two target lands", effects.Land()).WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TuckLegalCardTargetsOnTop(ctx)
		},
	})
}
