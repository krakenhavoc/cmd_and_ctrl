package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shifting Grift — Sorcery {U}{U} (Outlaws of Thunder Junction, #4468):
//
//	"Spree (Choose one or more additional costs.)
//	 + {2} — Exchange control of two target creatures.
//	 + {1} — Exchange control of two target artifacts.
//	 + {1} — Exchange control of two target enchantments."
//
// Switcheroo's exchange (CR 701.12) three times over, one Spree bullet
// (CR 702.172) each. Each exchange is all or nothing (CR 701.12a): if
// either of a bullet's two targets is illegal as it resolves, that pair
// stays put, and the other bullets still happen. Neither half has a
// duration. The same permanent may be picked by two bullets (an
// artifact creature, say); the bullets resolve in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "763d60ea-a0ad-4f13-b8a7-aa1515283266",
		Name:         "Shifting Grift",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Exchange control of two target creatures.", "{2}",
				TargetCreature("two target creatures").WithCount(2, 2), shiftingGriftExchange("creatures")),
			SpreeModeDoing("Exchange control of two target artifacts.", "{1}",
				TargetPermanent("two target artifacts", Artifact()).WithCount(2, 2), shiftingGriftExchange("artifacts")),
			SpreeModeDoing("Exchange control of two target enchantments.", "{1}",
				TargetPermanent("two target enchantments", Enchantment()).WithCount(2, 2), shiftingGriftExchange("enchantments")),
		),
	})
}

// shiftingGriftExchange is one bullet: exchange control of its two
// targets, only when both are still legal.
func shiftingGriftExchange(what string) func(*game.StackItem, *Context, int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		var legal []game.TargetRef
		for _, t := range ctx.ModeTargets(occ) {
			if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
				legal = append(legal, t)
			}
		}
		if len(legal) != 2 {
			return nil
		}
		return ExchangeControl{
			A:     legal[0].ID,
			B:     legal[1].ID,
			Label: "Shifting Grift — exchange control of two " + what,
		}.Apply(ctx)
	}
}
