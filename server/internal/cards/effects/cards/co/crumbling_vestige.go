package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crumbling Vestige — Land (Oath of the Gatewatch, S82 OLD #5545):
//
//	"This land enters tapped.
//	 When this land enters, add one mana of any color.
//	 {T}: Add {C}."
//
// The enters trigger is not a mana ability (CR 605.1b): it uses the stack and
// adds its mana as it resolves, Outcaster Trailblazer's pick of all five
// colours. The tapped entry is the land's own replacement (Bad River).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c1fe86a-3696-4634-bc26-cd45846529ec",
		Name:         "Crumbling Vestige",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Crumbling Vestige — add one mana of any color",
				effects.Do(effects.AddMana{Produced: "{W|U|B|R|G}"})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
