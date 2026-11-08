package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tempest Djinn — Creature — Djinn {U}{U}{U}, 0/4 (Foundations, #4465):
//
//	"Flying
//	 This creature gets +1/+0 for each basic Island you control."
//
// Molten Man's layer 7c self-static, counting power only and only basic
// Islands: a land you control with the Basic supertype and the Island
// subtype, both read off its effective characteristics, so a nonbasic
// Island (a shockland) is not counted and a basic land made an Island is.
// "Island" is read through the Djinn's text (CR 612.2). Flying rides
// PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b94c233e-11d3-4556-a1da-bd193e58b663",
		Name:            "Tempest Djinn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+0 for each basic Island you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				island := source.TextSubtype("Island")
				c.Power += permanentsYouControl(g, source, func(o *game.Card) bool {
					return o.IsLand() && o.HasSupertype("Basic") && o.HasSubtype(island)
				})
			},
		}},
	})
}
