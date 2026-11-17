package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant Tortoise — Creature — Turtle {1}{U}, 1/1 (Arabian Nights,
// #5641):
//
//	"This creature gets +0/+3 as long as it's untapped."
//
// A self-only layer 7c pump read off the Tortoise's tapped status, the
// way Illusion Spinners reads it for hexproof: it is a 1/4 at rest and a
// 1/1 once tapped (attacking without vigilance, or tapped by an effect).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2dd50d7f-941f-4deb-a15c-ee2357844c35",
		Name:         "Giant Tortoise",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "This creature gets +0/+3 as long as it's untapped.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && !target.Tapped
			},
			Apply: pumpBy(0, 3),
		}},
	})
}
