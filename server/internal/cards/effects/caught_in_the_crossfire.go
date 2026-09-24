package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caught in the Crossfire — Instant {R}{R}:
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Caught in the Crossfire deals 2 damage to each outlaw
//	   creature. (Assassins, Mercenaries, Pirates, Rogues, and
//	   Warlocks are outlaws.)
//	 + {1} — Caught in the Crossfire deals 2 damage to each non-outlaw
//	   creature."
//
// Spree proof card #4 (CR 702.172a, ADR 0065's 2026-09-23 amendment):
// both bullets are UNTARGETED mass-damage sweeps over a battlefield
// filter, proving Spree composes with a mode that never opens a
// targeting prompt at all — choosing both bullets is "deal 2 damage
// to each creature" for two mana total, split across the outlaw
// batch b40Outlaw already defines for Shoot the Sheriff.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d262f3e1-b5bc-4dab-86c2-cd89c2419e54",
		Name:         "Caught in the Crossfire",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Caught in the Crossfire deals 2 damage to each outlaw creature.", "{1}", nil,
				func(item *game.StackItem, ctx *Context, occ int) error {
					return damageEachMatching(ctx, And(Creature(), b40Outlaw()), 2)
				}),
			SpreeModeDoing("Caught in the Crossfire deals 2 damage to each non-outlaw creature.", "{1}", nil,
				func(item *game.StackItem, ctx *Context, occ int) error {
					return damageEachMatching(ctx, And(Creature(), Not(b40Outlaw())), 2)
				}),
		),
	})
}
