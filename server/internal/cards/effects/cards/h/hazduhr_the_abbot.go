package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hazduhr the Abbot — Legendary Creature — Human Cleric {3}{W}{W}, 2/5:
//
//	"{X}, {T}: The next X damage that would be dealt this turn to target
//	 white creature you control is dealt to Hazduhr instead."
//
// ADR 0108 §9 (#1905): a charged redirection (CR 615.7) of X, fixed as
// the ability was activated, from the target to Hazduhr as it is now.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35844d8b-8f68-4248-b921-e766757a9e26",
		Name:         "Hazduhr the Abbot",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: The next X damage that would be dealt this turn to target white creature you control is dealt to Hazduhr instead.",
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Targets: effects.TargetCreature("target white creature you control", effects.OfColor("W"), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if ctx.X() < 1 {
					return nil
				}
				return effects.RedirectDamage{Protect: effects.ShieldTheTarget, Amount: ctx.X(), To: effects.RedirectToThis}.Apply(ctx)
			},
		}},
	})
}
