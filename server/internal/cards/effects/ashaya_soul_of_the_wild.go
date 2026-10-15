package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashaya, Soul of the Wild — Legendary Creature — Elemental {3}{G}{G},
// */* (EDHREC top 1,000, #3654):
//
//	"Ashaya's power and toughness are each equal to the number of lands
//	 you control.
//	 Nontoken creatures you control are Forest lands in addition to
//	 their other types. (They're still affected by summoning sickness.)"
//
// Two layers over the same board. Layer 4 adds the Land type and the
// Forest subtype to your nontoken creatures (Ashaya included), which the
// engine already turns into a {T}: Add {G} ability from the land type,
// the way Urborg's Swamps work; the creatures keep their summoning
// sickness. Layer 7a is the characteristic-defining P/T, counted after
// the type change, so every nontoken creature you control, Ashaya
// herself among them, is one of the lands she counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "162572f2-1757-42e9-bd97-e6bd9a762c0e",
		Name:         "Ashaya, Soul of the Wild",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7A_CDA,
				Label:     "Ashaya's power and toughness are each equal to the number of lands you control.",
				AppliesTo: selfOnly,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := b10LandsControlled(g, source.Controller)
					c.Power, c.Toughness = n, n
				},
			},
			{
				Layer: game.Layer4Type,
				Label: "Nontoken creatures you control are Forest lands in addition to their other types.",
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && !target.IsToken() && target.Controller == source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !slices.Contains(c.Types, "Land") {
						c.Types = append(c.Types, "Land")
					}
					if !slices.Contains(c.Subtypes, "Forest") {
						c.Subtypes = append(c.Subtypes, "Forest")
					}
				},
			},
		},
	})
}
