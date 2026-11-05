package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collective Drone — Artifact Creature — Borg {1}{B}, */2 (Star Trek):
//
//	Collective Drone's power is equal to the number of artifact creatures
//	you control.
//
// A layer 7a characteristic-defining ability (CR 604.3) that counts itself,
// so on the battlefield it is always at least 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9806ba7-8af4-4e8e-9238-7bc40c6bee0b",
		Name:         "Collective Drone",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Collective Drone's power is equal to the number of artifact creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for i := range g.Battlefield.Cards {
					p := &g.Battlefield.Cards[i]
					if p.Controller == source.Controller && p.IsArtifact() && p.IsCreature() {
						n++
					}
				}
				c.Power = n
			},
		}},
	})
}
