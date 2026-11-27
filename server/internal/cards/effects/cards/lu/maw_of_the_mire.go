package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maw of the Mire — {4}{B} Sorcery (Innistrad, S82 OLD b447, #5572):
//
//	"Destroy target land. You gain 4 life."
//
// The life is gained even if the land was not destroyed (it was indestructible
// or regenerated), but not when the target has become illegal: the whole spell
// is then countered on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44e9030d-dd49-46b4-a0eb-75a49b6638a2",
		Name:         "Maw of the Mire",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 4}.Apply(ctx)
		},
	})
}
