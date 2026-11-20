package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earth Servant — Creature — Elemental {5}{R}, 4/4 (Magic 2011, #5579):
//
//	"This creature gets +0/+1 for each Mountain you control."
//
// Flourishing Bloom-Kin's self pump with Granite Grip's count: a layer 7c
// static on the Servant itself, the Mountains counted on every pass among
// the permanents its controller controls. "Mountain" is a word of the
// card's text (CR 612.2), so a text change counts another land type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ce30379-3f07-4b94-b2d6-a89b4763971b",
		Name:         "Earth Servant",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +0/+1 for each Mountain you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Toughness += mountainsTheAurasControllerControls(g, source)
			},
		}},
	})
}
