package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Upriser Renegade — Creature — Human Samurai {1}{R}, 1/3 (Kamigawa: Neon
// Dynasty):
//
//	"This creature gets +2/+0 for each other modified creature you control.
//	 (Equipment, Auras you control, and counters are modifications.)"
//
// A layer 7c static on itself, counting live (b07IsModified is safe inside
// a layer recompute).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f9b3ef8-e0ce-4906-96b3-a7727d7cc662",
		Name:         "Upriser Renegade",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +2/+0 for each other modified creature you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += 2 * permanentsYouControl(g, source, func(other *game.Card) bool {
					return other.InstanceID != source.InstanceID && other.IsCreature() && b07IsModified(g, *other)
				})
			},
		}},
	})
}
