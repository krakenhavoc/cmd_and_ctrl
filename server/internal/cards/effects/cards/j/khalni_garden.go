package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Khalni Garden — Land (EDHREC rank 3175):
//
//	"This land enters tapped.
//	 When this land enters, create a 0/1 green Plant creature token.
//	 {T}: Add {G}."
//
// A Forest that brings a chump blocker. The tapped entry is the real
// CR 614 self-replacement; the Plant is an enters trigger on the
// stack, so it can be responded to; the mana ability is a plain
// green tap. The Plant is Avenger of Zendikar's 0/1 green Plant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2d5ba45-8674-4428-89db-c2bbbf0bf5c5",
		Name:         "Khalni Garden",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Khalni Garden — create a 0/1 green Plant", effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 green Plant"), N: 1})),
		},
	})
}
