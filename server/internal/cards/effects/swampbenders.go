package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swampbenders — Creature — Human Druid Ally {4}{G}{G}, */* (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"Swampbenders's power and toughness are each equal to the number of
//	 Swamps on the battlefield.
//	 Lands you control are Swamps in addition to their other types."
//
// The layer-7a characteristic-defining ability counts every Swamp on
// the battlefield, under any controller, through the layered type line;
// the layer-4 grant runs first, so each land Swampbenders' controller
// controls counts. A land that is a Swamp has "{T}: Add {B}." from its
// land type, and the grant leaves the land's other types, names and
// abilities as they were. Land cards off the battlefield are untouched
// (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "feb00c6d-ff43-49bc-84d5-57e2d4651258",
		Name:         "Swampbenders",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Power and toughness are each equal to the number of Swamps on the battlefield.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7A_CDA,
				AppliesTo: selfOnly,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := avB27SwampsOnTheBattlefield(g, source.TextSubtype("Swamp"))
					c.Power = n
					c.Toughness = n
				},
			},
			avB27LandsYouControlAreAlso("Swamp"),
		},
	})
}
