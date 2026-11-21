package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Utter End — Instant {2}{W}{B} (EDHREC rank 1684):
//
//	"Exile target nonland permanent."
//
// Orzhov's unconditional answer: an instant, an exile, anything but
// a land — Anguished Unmaking without the life payment. Exile rather
// than destroy, so indestructible does not save the target and no
// dies-trigger fires, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cba94732-1f1f-4ffd-9aba-45db990043fa",
		Name:         "Utter End",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: exiles a permanent
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			return effects.ExileTarget{Target: legal[0].ID}.Apply(ctx)
		},
	})
}
