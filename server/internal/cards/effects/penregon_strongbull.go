package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a4e34111-aefd-4c61-a6e3-ac8895799c3f",
		Name:         "Penregon Strongbull",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice an artifact: This creature gets +1/+1 until end of turn and deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeN(1, "an artifact", Artifact())),
			Effect: Do(
				Instruction(func(ctx *Context) error {
					if !onBattlefield(ctx.Game, ctx.Source()) {
						return nil
					}
					return BoostUntilEOT{Target: ctx.Source(), Power: 1, Toughness: 1, Label: "Penregon Strongbull — +1/+1"}.Apply(ctx)
				}),
				DamageEachOpponent{N: 1},
			),
		}},
	})
}
