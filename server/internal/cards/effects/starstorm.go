package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starstorm — Instant {X}{R}{R}:
//
//	"Starstorm deals X damage to each creature.
//	 Cycling {3}"
//
// Mass damage, one damage instance (damageEachMatching); a creature whose
// toughness is above X survives. Cycling is the shared CR 702.29a ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34d112c3-cc2d-4428-bd7f-568bda61bd30",
		Name:         "Starstorm",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}},
		Activated:    []ActivatedAbility{Cycling("{3}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Creature(), ctx.X())
		},
	})
}
