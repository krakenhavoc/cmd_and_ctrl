package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sumala Rumblers — Creature — Wurm {2}{G/W}{G/W}, */4 (Ravnica: Clue Edition, #5929): "Its power is equal to the number of creatures you control. Myriad."
//
// A layer 7a CDA over the controller's creatures, itself included; myriad is GrantedMyriad.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09189f67-b33f-4775-bfd5-862154da7526",
		Name:         "Sumala Rumblers",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{powerEqualToCreaturesYouControl()},
		Triggered:    []game.TriggeredAbility{GrantedMyriad("Sumala Rumblers")},
	})
}

// powerEqualToCreaturesYouControl is the layer 7a CDA over the
// controller's creatures (PowerEqualToArtifactsYouControl's shape).
func powerEqualToCreaturesYouControl() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Power is equal to the number of creatures you control.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			for _, p := range g.Battlefield.Cards {
				if p.Controller == source.Controller && p.IsCreature() {
					n++
				}
			}
			c.Power = n
		},
	}
}
