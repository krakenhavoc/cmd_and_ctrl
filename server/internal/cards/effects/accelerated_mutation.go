package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Accelerated Mutation — Instant {3}{G}{G} (Scourge):
//
//	"Target creature gets +X/+X until end of turn, where X is the greatest
//	 mana value among permanents you control."
//
// X is read once, as the spell resolves (scgB1GreatestManaValue), and the
// boost is fixed from then on: a bigger permanent arriving later in the
// turn does not grow it. The X is the text's, not a cost, so the declared
// purpose carries no amount.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8a9c279-0f11-4a22-8651-f9caf013ca3c",
		Name:         "Accelerated Mutation",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := scgB1GreatestManaValue(ctx)
			return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Accelerated Mutation — +X/+X until end of turn"}.Apply(ctx)
		},
	})
}
