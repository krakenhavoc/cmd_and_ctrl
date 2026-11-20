package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cumber Stone — Artifact {3}{U} (Conflux, #5587):
//
//	"Creatures your opponents control get -1/-0."
//
// Waker of Waves's static on an artifact: an opponents-only layer 7c
// modifier (CR 613.4c) on power alone, read continuously, so a creature
// that changes hands gains or loses it at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49199997-65cd-4cb9-bdd6-74c93c1f6b5a",
		Name:         "Cumber Stone",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Creatures your opponents control get -1/-0.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller != source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power--
			},
		}},
	})
}
