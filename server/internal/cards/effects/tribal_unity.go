package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tribal Unity — Instant {X}{2}{G}:
//
//	"Creatures of the creature type of your choice get +X/+X until end
//	 of turn."
//
// Chosen as the spell resolves (#2382). It pumps EVERY creature of the
// type, yours and everyone else's, and the set is fixed when the answer
// arrives (CR 611.2c): a creature that enters later this turn gets
// nothing. A changeling is of every type and is always in the set.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b4ee935-4474-4757-9fab-978a2e1e72a0",
		Name:         "Tribal Unity",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeThen(ctx.Game, item.Controller, item.SourceCardID, "Tribal Unity — choose a creature type",
				func(g *game.Game, t string) error {
					if t == "" {
						return nil
					}
					c := NewContext(g, item)
					return BoostUntilEOT{
						Match:     OfCreatureType(t),
						Power:     c.X(),
						Toughness: c.X(),
						Label:     "Tribal Unity — +X/+X",
					}.Apply(c)
				})
			return nil
		},
	})
}
