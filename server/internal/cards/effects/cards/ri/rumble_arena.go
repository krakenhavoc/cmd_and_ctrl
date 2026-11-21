package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rumble Arena — Land (Avatar: The Last Airbender, #4459):
//
//	"Vigilance
//	 When this land enters, scry 1.
//	 {T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// Crystal Grotto with vigilance printed on it. Vigilance is printed data
// (CR 702.20): it does nothing while the land is not a creature, and the
// moment an effect animates it (earthbend, a manland effect) it attacks
// without tapping. The scry is a real enters trigger with a response
// window.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83f2470e-2114-4cd9-ab1a-21b76e9cdf63",
		Name:            "Rumble Arena",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Scry: 1},
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []effects.ManaAbility{
			effects.PainlessColorless(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rumble Arena — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
