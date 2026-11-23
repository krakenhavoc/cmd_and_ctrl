package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gristleback — Creature — Boar Beast {2}{G}, 2/2 (Guildpact, #5599):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)
//	 Sacrifice this creature: You gain life equal to this creature's
//	 power."
//
// The shared Bloodthirst entry replacement, and a sacrifice whose amount
// is the Boar's power as it last existed on the battlefield (CR 608.2h,
// the same read as Voldaren Thrillseeker's fling), its counter and any
// pump included. A power of 0 or less gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36f4d359-f46c-4da4-a8c2-1fe1799ac3f7",
		Name:         "Gristleback",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Bloodthirst(1, "Gristleback")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: You gain life equal to this creature's power.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok || info.Power <= 0 {
					return nil
				}
				return effects.GainLife{Player: item.Controller, Amount: info.Power}.Apply(ctx)
			},
		}},
	})
}
