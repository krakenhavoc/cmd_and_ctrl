package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e277ed07-2113-4465-a45c-a64490814c1d",
		Name:         "Impossible Choice",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Surgical Separation — Destroy target creature. Its controller creates two 1/1 red Officer creature tokens.",
				TargetCreature("target creature")),
				game.Purpose{
					Targets: game.ForTargets(RemovalReturning(0, game.TargetReturn{CreatureTokens: 2, TokenPower: 1, TokenToughness: 1})),
					Answers: game.AnswerRemove,
				}),
			Mode("Remain Fused — You draw two cards and each player loses 2 life."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				return impossibleChooseSeparation(item, ctx)
			case ctx.HasMode(1):
				if err := (DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
					return err
				}
				for _, id := range tablePlayers(ctx) {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -2); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}

func impossibleChooseSeparation(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	target := item.Targets[0].ID
	controller, ok := controllerOfTarget(ctx, target)
	if !ok {
		return nil
	}
	if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Controller: controller, Template: TokenCard("1/1 red Officer"), N: 2}.Apply(ctx)
}
