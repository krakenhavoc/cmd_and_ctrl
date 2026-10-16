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
// Forest stays a Forest, just gains the Artifact stamp. Its ruling:
// spells and cards in other zones "aren't permanents, so those spells
// and cards don't become artifacts."
//
// The third line is the player static #1600 built for Chromatic
// Orrery, covering every player rather than the controller: the engine
// reads it at every payment any player makes (game/spend_any_color.go,
// CR 609.4b).
//
// The second line "makes everything, in every zone of the game,
// colorless" (its ruling), and it is two statics: a layer 5 static
// over every permanent (#3172), and one over every spell on the stack
// and every card in a hand, library, graveyard, exile or the command
// zone (#3862, game/zone_layers.go). A colourless card in a hand is
// colourless to "protection from red", to "a colorless spell" and to a
// "red card" search alike. A commander's colour identity is not its
// colour (CR 903.4) and is unaffected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ae1f2ab5-c6a5-4d49-a746-3cb4668bf805",
		Name:          "Mycosynth Lattice",
		Completeness:  CompletenessFull,
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
				Layer:        game.Layer5Color,
				Label:        "All cards that aren't on the battlefield and spells are colorless.",
				AffectsZones: OffTheBattlefield(),
				AppliesTo: func(_ *game.Card, _ *game.Game, _ *game.Card) bool {
					return true // all spells and cards
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
