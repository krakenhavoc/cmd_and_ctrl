package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snow Villiers — Legendary Creature — Human Rebel Monk {2}{W}, */3
// (Final Fantasy, #4462):
//
//	"Vigilance
//	 Snow Villiers's power is equal to the number of creatures you
//	 control."
//
// Ironroot Warlord's characteristic-defining ability (layer 7a, Snow
// himself counted), and vigilance rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01cdaf69-f53a-4735-92f2-ef149de4d5c9",
		Name:            "Snow Villiers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			jmpB5PowerIsCreaturesYouControl("Snow Villiers's power is equal to the number of creatures you control."),
		},
	})
}
