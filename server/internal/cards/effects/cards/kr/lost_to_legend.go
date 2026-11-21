package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lost to Legend — Instant {W}{W} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Put target nonland historic permanent into its owner's library fourth
//	 from the top. (Artifacts, legendaries, and Sagas are historic.)"
//
// A nonland permanent that is an artifact, is legendary, or is a Saga, each read
// through the layers (CR 700.6). It goes fourth from the top of its owner's
// library; a shorter library takes it on the bottom.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a92edf30-6c7f-481c-855e-2a7dc4963ea7",
		Name:         "Lost to Legend",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("target nonland historic permanent",
			effects.Nonland(), effects.Or(effects.Artifact(), effects.Legendary(), effects.HasSubtype("Saga"))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.PutIntoLibrary{Card: t.ID, Depth: 4}.Apply(ctx)
		},
	})
}
