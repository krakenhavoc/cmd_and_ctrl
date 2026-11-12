package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kird Ape — Creature — Ape {R}, 1/1 (Arabian Nights, #5641):
//
//	"This creature gets +1/+2 as long as you control a Forest."
//
// Drover of the Mighty's conditional pump with a land subtype: any
// permanent you control with the Forest subtype counts, a nonbasic
// Forest or a land made a Forest by an effect included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbbc3acb-c917-44ff-ac6f-9dd6ebe3f4ad",
		Name:         "Kird Ape",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+2 as long as you control a Forest.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Forest")),
			Apply:     pumpBy(1, 2),
		}},
	})
}
