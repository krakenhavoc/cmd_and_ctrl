package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outbreak — Sorcery {3}{B}:
//
//	"You may discard a Swamp card rather than pay this spell's mana
//	 cost.
//	 Choose a creature type. All creatures of that type get -1/-1 until
//	 end of turn."
//
// ADR 0135 §2 (#2412): Snag's discard price. The type is chosen as the
// spell resolves (#2382), and the set of creatures is fixed when the
// answer arrives (CR 611.2c): one that enters later this turn is not
// shrunk. Every creature of the type is affected, yours included, and a
// changeling is of every type.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:         "05079cee-b190-42d7-b59f-130cb545cab7",
		Name:             "Outbreak",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{DiscardInstead("a Swamp card", HasSubtype("Swamp"))},
		// ADR 0126 owner decision 2: a -1/-1 sweep of one chosen type,
		// so partial.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeThen(ctx.Game, item.Controller, item.SourceCardID, "Outbreak — choose a creature type",
				func(g *game.Game, t string) error {
					if t == "" {
						return nil
					}
					return BoostUntilEOT{
						Match:     OfCreatureType(t),
						Power:     -1,
						Toughness: -1,
						Label:     "Outbreak — -1/-1",
					}.Apply(NewContext(g, item))
				})
			return nil
		},
	})
}
