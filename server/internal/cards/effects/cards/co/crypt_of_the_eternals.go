package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crypt of the Eternals — Land (Hour of Devastation, S82 OLD batch 241, #5534):
//
//	"When this land enters, you gain 1 life.
//	 {T}: Add {C}.
//	 {1}, {T}: Add {U}, {B}, or {R}."
//
// Two mana abilities: the colourless one, and a filter that pays {1} and taps
// for one mana of a colour. Both are real mana abilities (CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc78776b-822b-4f11-8982-0805a25a9d36",
		Name:         "Crypt of the Eternals",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "Add {C}"},
			{Cost: effects.ManaAbilityCost{Tap: true, Mana: "{1}"}, Produced: "{U|B|R}", Label: "{1}, {T}: Add {U}, {B}, or {R}"},
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Crypt of the Eternals — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
