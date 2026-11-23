package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Generous Gift — Instant {2}{W}:
//
//	"Destroy target permanent. Its controller creates a 3/3 white
//	Elephant creature token."
//
// White's Beast Within, printing-for-printing identical in shape.
// See beast_within.go for why the controller is read before the
// destroy rather than after.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fae37e28-e137-4177-b973-fa8b4dd8f409",
		Name:         "Generous Gift",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		// #2679: what the target's controller is given back.
		Purpose: game.Purpose{Targets: game.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3})), Answers: game.AnswerRemove}, // #4734: destroys a permanent
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
				Template:   effects.TokenCard("3/3 green Elephant"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
