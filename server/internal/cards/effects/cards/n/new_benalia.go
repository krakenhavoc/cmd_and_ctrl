package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// New Benalia — Land (Future Sight, #5594):
//
//	"This land enters tapped.
//	 When this land enters, scry 1. (Look at the top card of your
//	 library. You may put that card on the bottom.)
//	 {T}: Add {W}."
//
// A one-colour Temple: the CR 614 self-replacement for entering tapped,
// the scry as the land's own trigger on the stack, and a plain {W} mana
// ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e743fbf-b5b6-4176-a4f2-6933f521f2fe",
		Name:         "New Benalia",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("New Benalia — scry 1", effects.Do(effects.Scry{N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
