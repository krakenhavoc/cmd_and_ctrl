package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Renata, Called to the Hunt — Legendary Enchantment Creature —
// Demigod {2}{G}{G}, */3 (EDHREC rank 2843):
//
//	"Renata's power is equal to your devotion to green. (Each {G} in
//	 the mana costs of permanents you control counts toward your
//	 devotion to green.)
//	 Each other creature you control enters with an additional
//	 +1/+1 counter on it."
//
// The green Demigod. The power is a Layer 7a CDA (Daxos's shape) set
// to devotionTo(controller, "G") on every recompute — Renata's own
// {G}{G} counts, so she is at least a 2/3 while she is on the
// battlefield; the printed "*" is 0 in the card data and the CDA
// overrides it. The counter is Arwen's CR 614 entry replacement
// (b17OtherCreaturesYouControlEnterWithCounters) with a fixed one
// counter — cast, reanimated, fetched or flickered creature cards
// all get it, and Hardened Scales and Doubling Season apply to it.
//
// Sandbox simplification, declared: creature TOKENS get nothing.
// Token creation skips the CR 614 zone-move pipeline (the Urabrask
// the Hidden gap), so no entry replacement sees a token — Arwen's
// declared gap, shared. Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "6761b077-7a89-42c9-93ab-675fe4231564",
		Name:         "Renata, Called to the Hunt",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Creature tokens you create don't get the extra +1/+1 counter — only creature cards entering the battlefield do."},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = devotionTo(g, source.Controller, "G")
			},
		}},
		Replacements: []game.ReplacementEffect{
			b27OtherCreaturesYouControlEnterWithACounter("Renata, Called to the Hunt: enters with an additional +1/+1 counter"),
		},
	})
}
