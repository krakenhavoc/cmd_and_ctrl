package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Militant Inquisitor — Creature — Human Cleric {2}{W}, 2/3 (Foundations
// Jumpstart):
//
//	"This creature gets +1/+0 for each Equipment you control."
//
// A layer 7c self-modify whose count is re-read on every layer pass
// (Adelbert Steiner's shape): every Equipment its controller controls
// counts, attached to it or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1e326e8-6e9e-4dbd-8c2f-0bef90db2bb7",
		Name:         "Militant Inquisitor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+0 for each Equipment you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += permanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype("Equipment") })
			},
		}},
	})
}
