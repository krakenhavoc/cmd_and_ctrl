package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firemind Vessel — Artifact {4}:
//
//	"This artifact enters tapped.
//	 {T}: Add two mana of different colors."
//
// The plain "two mana of different colors" rock (#2558): one pick of
// two DIFFERENT colours (DifferentColors(2)), so it can pay {W}{U} and
// never {U}{U}. Clicked, it offers the ten pairs; asked through the
// mana_pick prompt, it asks for the second colour with the first struck
// out; the auto-tapper books the pair a cost needs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8afd7b0-7cd0-4caf-ac43-a0e24a93ca09",
		Name:         "Firemind Vessel",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.DifferentColors(2),
			Label:    "Add two mana of different colors",
		}},
	})
}
