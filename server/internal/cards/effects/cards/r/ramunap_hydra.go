package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ramunap Hydra — {3}{G} Creature — Snake Hydra 3/3 (Hour of Devastation):
//
//	"Reach, vigilance, trample
//	 This creature gets +1/+1 as long as you control a Desert.
//	 This creature gets +1/+1 as long as there is a Desert card in your
//	 graveyard."
//
// Two separate statics, as printed: with both a Desert and a Desert card it is
// +2/+2.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2dadaf5-eaa1-4c43-9e21-bb2334ceb918",
		Name:            "Ramunap Hydra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "trample"},
		Static: []game.StaticAbility{
			{
				Label:     "This creature gets +1/+1 as long as you control a Desert.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.OldB240SelfWhile(effects.OldB240YouControlADesert),
				Apply:     effects.PumpBy(1, 1),
			},
			{
				Label:     "This creature gets +1/+1 as long as there is a Desert card in your graveyard.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.OldB240SelfWhile(effects.OldB240DesertCardInYourGraveyard),
				Apply:     effects.PumpBy(1, 1),
			},
		},
	})
}
