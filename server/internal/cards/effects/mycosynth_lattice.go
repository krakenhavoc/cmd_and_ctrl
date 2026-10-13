package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mycosynth Lattice — Artifact {6}:
//
//	"All permanents are artifacts in addition to their other types.
//	 All cards that aren't on the battlefield, spells, and permanents
//	 are colorless.
//	 Players may spend mana as though it were mana of any color."
//
// First S16 Layer-4 catalog card. One static ability that walks
// every battlefield permanent on each recompute pass and appends
// "Artifact" to its effective types — idempotent so the printed
// Forest stays a Forest, just gains the Artifact stamp.
//
// The third line is the player static #1600 built for Chromatic
// Orrery, covering every player rather than the controller: the engine
// reads it at every payment any player makes (game/spend_any_color.go,
// CR 609.4b).
//
// The second line is half built (#3172, cav-h1): a layer 5 static clears
// the colors of every PERMANENT. The layer engine does not recompute a
// spell on the stack or a card in a hidden zone, so "all cards that
// aren't on the battlefield, spells" stays coloured, which is why the
// card keeps its caveat. The half is the safe one to ship: a permanent
// that loses its colors is the Lattice's own text, and a spell still
// being coloured only matters to a "protection from red" or "colorless
// spell" reader. The rest waits for a colour pass that reaches the
// stack and the other zones.
func init() {
	Register(Spec{
		OracleID:      "ae1f2ab5-c6a5-4d49-a746-3cb4668bf805",
		Name:          "Mycosynth Lattice",
		Completeness:  CompletenessCaveats,
		Caveats:       []string{"Permanents are colorless, but spells on the stack and cards in hand, graveyard or exile keep their colors."},
		AnyColorSpend: PlayersMaySpendManaAsAnyColor(),
		Static: []game.StaticAbility{
			{
				Layer: game.Layer5Color,
				Label: "All permanents are colorless.",
				AppliesTo: func(_ *game.Card, _ *game.Game, _ *game.Card) bool {
					return true // all permanents
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Colors = nil
				},
			},
			{
				Layer: game.Layer4Type,
				Label: "All permanents are artifacts in addition to their other types.",
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
