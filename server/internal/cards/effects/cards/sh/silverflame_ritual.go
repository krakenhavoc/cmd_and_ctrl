package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silverflame Ritual — Sorcery {3}{W} (Throne of Eldraine, #5516):
//
//	"Put a +1/+1 counter on each creature you control.
//	 Adamant — If at least three white mana was spent to cast this spell,
//	 creatures you control gain vigilance until end of turn."
//
// Basri's Solidarity's counters, then the adamant rider over the creatures
// the caster controls as it resolves, the set locked then (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "d511573b-35c7-4934-aa46-906e14101664",
		Name:           "Silverflame Ritual",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"W"},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.B11PutCounterOnEachCreatureYouControl(ctx.Game, ctx.Item)
			}),
			effects.EldB1IfAdamantSpent("W", effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"vigilance"},
				Label:    "Silverflame Ritual — vigilance",
			}),
		),
	})
}
