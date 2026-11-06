package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boneclub Berserker — Creature — Goblin Berserker {3}{R}, 2/4 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature gets +2/+0 for each other Goblin you control."
//
// Living Lies of Loki's static at +2/+0 per creature: a layer 7c count of
// the other Goblin creatures you control, read on every pass through the
// effective view (a changeling counts) and through the source's text
// (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a08f69f9-de48-46d6-8b75-31886159cf45",
		Name:         "Boneclub Berserker",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +2/+0 for each other Goblin you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				goblin := source.TextSubtype("Goblin")
				c.Power += 2 * permanentsYouControl(g, source, func(other *game.Card) bool {
					return other.InstanceID != source.InstanceID && other.IsCreature() && other.HasSubtype(goblin)
				})
			},
		}},
	})
}
