package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Master of Etherium — Artifact Creature — Vedalken Wizard {2}{U},
// */* (EDHREC rank 2161):
//
//	"Master of Etherium's power and toughness are each equal to the
//	 number of artifacts you control.
//	 Other artifact creatures you control get +1/+1."
//
// The affinity lord. The P/T is a layer 7a characteristic-defining
// ability that SETS both values to the artifact count on every
// recompute (b15ArtifactsControlled — post-layer types, so an
// animated or Lattice'd permanent counts, and the Master counts
// itself, as printed). The anthem is a layer 7c +1/+1 for every
// OTHER artifact creature under the same controller, read on the
// post-layer type line so a creature an effect made an artifact gets
// it too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "181e6b3e-c33e-49e7-acb8-67473289a856",
		Name:         "Master of Etherium",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7A_CDA,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := b15ArtifactsControlled(g, source.Controller)
					c.Power = n
					c.Toughness = n
				},
			},
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID != source.InstanceID &&
						target.Controller == source.Controller &&
						target.IsCreature() && target.IsArtifact()
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
		},
	})
}
