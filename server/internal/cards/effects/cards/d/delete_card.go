package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Delete — Sorcery {X}{R}{R} (Doctor Who):
//
//	"Delete deals X damage to each nonartifact creature and each player."
//
// One damage instance (CR 120.2): every nonartifact creature, then every
// player still in the game, from Delete.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01d0156f-c26a-4668-a3a4-15ee0790d1b6",
		Name:         "Delete",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.Not(effects.Artifact())), x); err != nil {
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
