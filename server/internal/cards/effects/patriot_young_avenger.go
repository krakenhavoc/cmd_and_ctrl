package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patriot, Young Avenger — Legendary Creature — Human Hero {2}{R/W}, 3/2
// (Marvel Super Heroes Commander, #4455):
//
//	"Prowess
//	 As long as Patriot is equipped, other creatures you control get +1/+0."
//
// Prowess is derived from the keyword by the engine (game/prowess.go).
// The static is a layer 7c modification whose condition is read on every
// recompute: an Equipment attached to Patriot turns it on, and losing
// the Equipment turns it off in the same pass (b41SourceIsEquipped, the
// reader Merry, Esquire of Rohan uses). An Aura on Patriot is not
// Equipment, so it does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6c44d62f-52de-446c-9824-6b474534a612",
		Name:            "Patriot, Young Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			Label:    "As long as Patriot is equipped, other creatures you control get +1/+0.",
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.InstanceID != source.InstanceID && b41SourceIsEquipped(g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
			},
		}},
	})
}
