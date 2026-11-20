package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "36f4d359-f46c-4da4-a8c2-1fe1799ac3f7",
		Name:         "Gristleback",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Bloodthirst(1, "Gristleback")},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: You gain life equal to this creature's power.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok || info.Power <= 0 {
					return nil
				}
				return GainLife{Player: item.Controller, Amount: info.Power}.Apply(ctx)
			},
		}},
	})
}
