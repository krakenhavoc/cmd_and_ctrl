package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nephalia Smuggler — Creature — Human Rogue {U}, 1/1 (S82 GAP, #5948):
//
//	"{3}{U}, {T}: Exile another target creature you control, then return
//	 that card to the battlefield under your control."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f96a2aa6-3259-4bb6-92cd-4d17215d04d0",
		Name:         "Nephalia Smuggler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}, {T}: Exile another target creature you control, then return that card to the battlefield under your control.",
			Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
					return nil
				}
				return effects.Flicker{Target: item.Targets[0].ID, Controller: ctx.Controller()}.Apply(ctx)
			},
		}},
	})
}
