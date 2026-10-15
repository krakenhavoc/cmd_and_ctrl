package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const lifestreamsBlessingX = "greatest_power"

func init() {
	Register(Spec{
		OracleID:     "7cbdba72-ed33-40c6-a9c0-4d65c280a4d0",
		Name:         "Lifestream's Blessing",
		Completeness: CompletenessFull,
		CastTimeValues: []game.CastTimeValue{{
			Key: lifestreamsBlessingX,
			Value: func(g *game.Game, item *game.StackItem) int {
				return b42GreatestPowerControlledBy(g, item.Controller)
			},
		}},
		SpecialActions: []game.SpecialAction{
			Foretell("{4}{G}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x, ok := ctx.CastValue(lifestreamsBlessingX)
			if !ok {
				// A game restored from before the value was recorded:
				// the best available answer is the board as it resolves.
				x = b42GreatestPowerControlledBy(ctx.Game, ctx.Controller())
			}
			if x < 0 {
				x = 0
			}
			if err := (DrawCards{Player: ctx.Controller(), N: x}).Apply(ctx); err != nil {
				return err
			}
			if item.CastFromZone != game.ZoneExile {
				return nil
			}
			return GainLife{Player: ctx.Controller(), Amount: 2 * x}.Apply(ctx)
		},
	})
}
