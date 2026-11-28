package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rainbow Vale — Land (Fallen Empires, S82 OLD b855, #5637):
//
//	"{T}: Add one mana of any color. An opponent gains control of this land at the
//	 beginning of the next end step."
//
// The mana is added at once (CR 605.3b) and the rider creates a delayed triggered ability
// (CR 603.7) for the next end step, whichever turn it is. It names the land as the object
// it is now, so a land that has left and come back is not given away (CR 400.7). The
// land's controller chooses the opponent as the trigger resolves, with no question when
// there is one. The control change has no duration, so it lasts indefinitely (CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76695b15-d0ba-41eb-85f1-52ba5d14b8ba",
		Name:         "Rainbow Vale",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color. An opponent gains control of this land at the beginning of the next end step.",
			Rider:    effects.S82OldB855RainbowValeRider,
		}},
	})
}
