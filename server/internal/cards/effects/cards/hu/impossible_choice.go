package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impossible Choice — Instant {2}{B} (Star Trek, #4973 batch 12):
//
//	"Choose one —
//	 • Surgical Separation — Destroy target creature. Its controller
//	   creates two 1/1 red Officer creature tokens.
//	 • Remain Fused — You draw two cards and each player loses 2 life."
//
// The two names are flavour words on the mode bullets; they have no
// rules meaning.
//
// SURGICAL SEPARATION gives the tokens to the creature's controller as
// it was just before the destroy (Beast Within's reading), so an
// indestructible creature still leaves its controller two Officers.
//
// REMAIN FUSED draws first, then each player, you included, loses 2
// life, in seat order from you. Life loss, not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e277ed07-2113-4465-a45c-a64490814c1d",
		Name:         "Impossible Choice",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Surgical Separation — Destroy target creature. Its controller creates two 1/1 red Officer creature tokens.",
				effects.TargetCreature("target creature")),
				game.Purpose{
					Targets: game.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 2, TokenPower: 1, TokenToughness: 1})),
					Answers: game.AnswerRemove,
				}),
			effects.Mode("Remain Fused — You draw two cards and each player loses 2 life."),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return impossibleChooseSeparation(item, ctx)
			case ctx.HasMode(1):
				if err := (effects.DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
					return err
				}
				for _, id := range effects.TablePlayers(ctx) {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -2); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}

func impossibleChooseSeparation(item *game.StackItem, ctx *effects.Context) error {
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
	return effects.CreateToken{Controller: controller, Template: effects.TokenCard("1/1 red Officer"), N: 2}.Apply(ctx)
}
