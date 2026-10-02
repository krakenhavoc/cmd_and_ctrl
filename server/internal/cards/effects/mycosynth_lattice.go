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
// Out of scope: the second line. Layer 5 can make every PERMANENT
// colorless, but the layer engine does not recompute a spell on the
// stack or a card in a hidden zone, so "all cards that aren't on the
// battlefield, spells, and permanents are colorless" would be half
// true — permanents colorless, the spell that becomes one coloured on
// the stack — which is the kind of partial rule a "protection from
// red" or a "spend this mana only on colorless spells" reads wrongly
// in both directions. It waits for a colour pass that reaches the
// stack and the other zones.
func init() {
	Register(Spec{
		OracleID:      "ae1f2ab5-c6a5-4d49-a746-3cb4668bf805",
		Name:          "Mycosynth Lattice",
		Completeness:  CompletenessCaveats,
		Caveats:       []string{"Nothing is made colorless — cards, spells and permanents keep their colors. The other two lines work."},
		AnyColorSpend: PlayersMaySpendManaAsAnyColor(),
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
