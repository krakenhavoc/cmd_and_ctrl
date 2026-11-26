package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kolaghan Forerunners — Creature — Human Berserker {2}{R}, */3 (Dragons of
// Tarkir, S82 old batch 331):
//
//	"Trample
//	 Kolaghan Forerunners's power is equal to the number of creatures you
//	 control.
//	 Dash {2}{R} (You may cast this spell for its dash cost. If you do, it
//	 gains haste, and it's returned from the battlefield to its owner's hand at
//	 the beginning of the next end step.)"
//
// Ironroot Warlord's characteristic-defining power (layer 7a, itself counted),
// printed trample, and the shared dash cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "07b6fbd3-03f9-4aef-929e-514c799d75b9",
		Name:             "Kolaghan Forerunners",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{2}{R}")},
		Static: []game.StaticAbility{
			effects.JmpB5PowerIsCreaturesYouControl("Kolaghan Forerunners's power is equal to the number of creatures you control."),
		},
	})
}
