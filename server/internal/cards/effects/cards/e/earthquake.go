package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earthquake — Sorcery {X}{R} (EDHREC rank 2846):
//
//	"Earthquake deals X damage to each creature without flying and
//	 each player."
//
// The original red sweeper, and the one that hits the table too. X
// rides the cast (ctx.X()); the creatures are snapshotted before the
// first point lands (damageEachMatching), "without flying" read
// post-layer so a granted flying keeps a creature out of it, and
// then every seated player — the caster included — takes X. Damage,
// not destruction: a creature that took lethal dies at the
// state-based sweep, so indestructible and prevention shields both
// work as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "9a40614b-50a3-422c-849e-53c8b7d3d204",
		Name:     "Earthquake",
		// ADR 0126 §6: a creature with flying is spared.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), x); err != nil {
					return err
				}
				for _, p := range ctx.Game.Seats {
					if p == nil || p.Eliminated {
						continue
					}
					if err := (effects.DealDamage{Source: ctx.Source(), Target: p.ID, Amount: x}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
