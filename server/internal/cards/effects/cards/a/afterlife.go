package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Afterlife — Instant {2}{W} (Commander 2014):
//
//	"Destroy target creature. It can't be regenerated. Its controller
//	 creates a 1/1 white Spirit creature token with flying."
//
// Beast Within's shape: the controller is read before the destroy, then
// handed the token. A creature that is indestructible is still not
// destroyed, but its controller still gets the Spirit (the instructions
// are independent). A target that left in response fizzles the spell
// (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c13e2b5-961a-4031-84b1-15bd19b94286",
		Name:         "Afterlife",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 1, TokenToughness: 1})), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, t.ID)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: controller, Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(ctx)
		},
	})
}
