package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stroke of Midnight — Instant {2}{W} (EDHREC rank 205):
//
//	"Destroy target nonland permanent. Its controller creates a 1/1
//	 white Human creature token."
//
// Generous Gift that can't hit a land and hands back a 1/1 instead
// of a 3/3 — a smaller drawback for a smaller target set. The shape
// is Beast Within's exactly: read the controller, destroy, then
// create under the victim. Nonland() is the whole difference in the
// target clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a107e48-3d50-4941-95b1-10f2b29a4245",
		Name:         "Stroke of Midnight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		// #2679: what the target's controller is given back.
		Purpose: game.Purpose{Targets: game.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 1, TokenToughness: 1})), Answers: game.AnswerRemove}, // #4734: destroys a permanent
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := effects.ControllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{
				Controller: controller,
				Template:   effects.TokenCard("1/1 white Human"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
