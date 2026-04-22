package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mycosynth Lattice — "All permanents are artifacts in addition to
// their other types." (plus other clauses out of S16 scope, see
// below.)
//
// First S16 Layer-4 catalog card. One static ability that walks
// every battlefield permanent on each recompute pass and appends
// "Artifact" to its effective types — idempotent so the printed
// Forest stays a Forest, just gains the Artifact stamp.
//
// Out of scope (deferred to S17 / S18 / S19):
//   - "All lands are every basic land type." Layer 4 type-add of
//     Plains/Island/Swamp/Mountain/Forest to every land — needs
//     the same parser hook but interacts with the S15 mana ability
//     pipeline. Sub-PR 4 ships only the artifact half.
//   - "Lands tap for any color." Replacement effect on the basic-
//     land synthetic mana ability output — S17 territory.
//   - "Players can't pay anything other than mana." Cost replacement;
//     S17 territory.
//
// The S16 sandbox stops at the type-add. The remaining clauses are
// documented so future implementers don't think they were missed.
func init() {
	Register(Spec{
		OracleID: "16ddc91a-4b7c-4d12-bd0a-1b2f4ce5d12e",
		Name:     "Mycosynth Lattice",
		Static: []game.StaticAbility{
			{
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return true // all permanents
				},
				Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
					for _, t := range c.Types {
						if t == "Artifact" {
							return // idempotent: already an artifact
						}
					}
					c.Types = append(c.Types, "Artifact")
				},
			},
		},
	})
}
