package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Callous Bloodmage — Creature — Vampire Warlock {2}{B}, 2/1 (Strixhaven:
// School of Mages, #5502):
//
//	"When this creature enters, choose one —
//	 • Create a 1/1 black and green Pest creature token with "When this token
//	   dies, you gain 1 life."
//	 • You draw a card and you lose 1 life.
//	 • Exile target player's graveyard."
//
// The Pest is the catalog's PestToken, dies trigger included. The draw is made
// before the life is lost, as printed. The third bullet's target is chosen as
// the trigger goes on the stack and only for that bullet (CR 700.2c); a
// graveyard that has changed by resolution is exiled as it then is.
//
// No simplification.
func init() {
	t := effects.WhenThisEnters("Callous Bloodmage — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeWithPurpose(effects.ModeDoing("Create a 1/1 black and green Pest creature token with \"When this token dies, you gain 1 life.\"", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.PestToken(), N: 1}.Apply(ctx)
			}), game.Purpose{Tokens: 1}),
		effects.ModeWithPurpose(effects.ModeSteps("You draw a card and you lose 1 life.", nil,
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 1}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
			})),
			game.Purpose{Draws: 1}),
		effects.ModeDoing("Exile target player's graveyard.", effects.TargetPlayer("target player"),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				tg, ok := effects.ModeTarget(ctx, occ)
				if !ok || tg.Kind != game.TargetPlayer {
					return nil
				}
				return effects.ExileGraveyardForEffect(ctx.Game, item, tg.ID)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "96c12c12-140a-462d-bff0-0554d192b1f4",
		Name:         "Callous Bloodmage",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
