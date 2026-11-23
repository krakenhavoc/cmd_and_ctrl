package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Penregon Strongbull — Creature — Minotaur {2}{R}, 2/3 (The Brothers' War,
// S82):
//
//	"{1}, Sacrifice an artifact: This creature gets +1/+1 until end of turn
//	 and deals 1 damage to each opponent."
//
// The artifact is a cost and may be the Strongbull itself only if it were
// one (it is not). The pump goes on only while the Strongbull is still on
// the battlefield; the damage is dealt by it as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4e34111-aefd-4c61-a6e3-ac8895799c3f",
		Name:         "Penregon Strongbull",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice an artifact: This creature gets +1/+1 until end of turn and deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
						return nil
					}
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Toughness: 1, Label: "Penregon Strongbull — +1/+1"}.Apply(ctx)
				}),
				effects.DamageEachOpponent{N: 1},
			),
		}},
	})
}
