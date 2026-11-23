package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Speed Demon — Legendary Creature — Demon {3}{B}{B}, 5/5:
//
//	"Flying, trample
//	 Start your engines!
//	 At the beginning of your end step, you draw X cards and lose X
//	 life, where X is your speed."
//
// ADR 0138 (#2122). X is read as the trigger resolves (CR 608.2h), off
// the controller's speed; a player with no speed draws nothing and
// loses nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c9fb22b4-dc22-4291-91f2-89c07cee7569",
		Name:            "The Speed Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("The Speed Demon — draw X cards and lose X life, where X is your speed",
				effects.Do(
					// The life loss waits for a prompt the draws ask (CR 608.2c, #4155). Speed
					// is read again for it: drawing does not change it.
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.DrawCards{N: effects.YourSpeed(ctx.Game, ctx.Item.Controller)}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.GainLife{Amount: -effects.YourSpeed(ctx.Game, ctx.Item.Controller)}.Apply(ctx)
					}),
				)),
		},
	})
}
