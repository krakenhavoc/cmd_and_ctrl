package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindspring Merfolk — Creature — Merfolk Wizard {U}, 1/1 (Aetherdrift,
// #4464):
//
//	"Exhaust — {X}{U}{U}, {T}: Draw X cards. Put a +1/+1 counter on each
//	 Merfolk creature you control. (Activate each exhaust ability only
//	 once.)"
//
// Boommobile's exhaust with a real announced X (the engine's once-only
// exhaust marker, keyed to the object, so a Merfolk back from a flicker
// may use it again). The draw is X cards; the counters then go on each
// Merfolk creature you control as the ability resolves, the Merfolk
// itself included if it is still there, read after the layers (a
// changeling counts) and through the ability's own text. The counter
// step waits for anything the draw asks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35142b01-4b8d-4335-a943-a3172fa04573",
		Name:         "Mindspring Merfolk",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "Exhaust — {X}{U}{U}, {T}: Draw X cards. Put a +1/+1 counter on each Merfolk creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Exhaust: true,
			Cost:    effects.Plus(effects.ManaCost("{X}{U}{U}"), effects.TapCost()),
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.DrawCards{N: ctx.X()}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.TrekB11PutCounterOnEachOfTypeYouControl("Merfolk")(ctx.Game, ctx.Item)
				}),
			),
		}},
	})
}
