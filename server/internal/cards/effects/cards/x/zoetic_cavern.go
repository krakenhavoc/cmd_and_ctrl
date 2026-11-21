package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zoetic Cavern — Land:
//
//	"{T}: Add {C}.
//	 Morph {2}"
//
// A land with morph is still cast face down as a spell for {3}, never
// played as a land (CR 702.37b); the face-down permanent is a 2/2
// creature that turns back into the land for {2}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3763de30-28e1-4689-a71c-07d2fea3a466",
		Name:         "Zoetic Cavern",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}")},
	})
}
