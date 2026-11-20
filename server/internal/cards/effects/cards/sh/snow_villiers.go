package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "01cdaf69-f53a-4735-92f2-ef149de4d5c9",
		Name:            "Snow Villiers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			effects.JmpB5PowerIsCreaturesYouControl("Snow Villiers's power is equal to the number of creatures you control."),
		},
	})
}
