package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akiri, Line-Slinger — Legendary Creature — Kor Soldier Ally {R}{W}, 0/3
// (Commander 2016, #5537):
//
//	"First strike, vigilance
//	 Akiri gets +1/+0 for each artifact you control.
//	 Partner (You can have two commanders if both have partner.)"
//
// Guidelight Synergist's layer-7c self pump, counted on every pass from
// the artifacts its controller controls now (artifact creatures and
// tokens included). Partner is a deck-building rule (CR 702.124) that
// internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3e0a0276-b640-41c7-95bd-ba3067f5c911",
		Name:            "Akiri, Line-Slinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance"},
		Static: []game.StaticAbility{{
			Label:     "Akiri gets +1/+0 for each artifact you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += artifactsControlledBy(g, source)
			},
		}},
	})
}
