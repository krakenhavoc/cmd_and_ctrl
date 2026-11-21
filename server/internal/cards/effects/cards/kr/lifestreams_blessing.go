package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lifestream's Blessing — Instant {4}{G}{G} (#3656):
//
//	"Draw X cards, where X is the greatest power among creatures you
//	 controlled as you cast this spell. If this spell was cast from
//	 exile, you gain twice X life.
//	 Foretell {4}{G}"
//
// Foretell is one special action. The life half reads the spell's cast
// zone, so any cast from exile gains the life, as printed.
//
// X is fixed AS YOU CAST (CR 601.2i): Spec.CastTimeValues stamps the
// greatest power among creatures the caster controls on the stack item
// once every cost is paid and before anyone can respond, and OnResolve
// reads it back with ctx.CastValue. A creature that dies, or grows, in
// response leaves the answer where it was. Layered power, counters
// included (b42GreatestPowerControlledBy); zero with no creature. A
// copy of the spell keeps the value (CR 707.10).
//
// No simplification.
// lifestreamsBlessingX is the key of the value the spell records as it is cast.
const LifestreamsBlessingX = "greatest_power"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cbdba72-ed33-40c6-a9c0-4d65c280a4d0",
		Name:         "Lifestream's Blessing",
		Completeness: effects.CompletenessFull,
		CastTimeValues: []game.CastTimeValue{{
			Key: LifestreamsBlessingX,
			Value: func(g *game.Game, item *game.StackItem) int {
				return effects.B42GreatestPowerControlledBy(g, item.Controller)
			},
		}},
		SpecialActions: []game.SpecialAction{
			effects.Foretell("{4}{G}"),
		},
		OnResolve: effects.ResolveSteps(
			// The life waits for a prompt the draws ask (CR 608.2c, #4155). X is the
			// cast-time value, read again for the second step.
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: lifestreamsBlessingValue(ctx)}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if ctx.Item.CastFromZone != game.ZoneExile {
					return nil
				}
				return effects.GainLife{Amount: 2 * lifestreamsBlessingValue(ctx)}.Apply(ctx)
			}),
		),
	})
}

// lifestreamsBlessingValue is X: the value recorded as the spell was cast.
func lifestreamsBlessingValue(ctx *effects.Context) int {
	x, ok := ctx.CastValue(LifestreamsBlessingX)
	if !ok {
		// A game restored from before the value was recorded:
		// the best available answer is the board as it resolves.
		x = effects.B42GreatestPowerControlledBy(ctx.Game, ctx.Controller())
	}
	return max(x, 0)
}
