package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lorehold Campus — Land (Secrets of Strixhaven Commander, #4834):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}.
//	 {4}, {T}: Scry 1."
//
// University Campus's shape with a scry: a nonbasic land with no basic
// types, so both mana colours are declared, and the scry is an ordinary
// activated ability whose cost taps the land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "45773715-3f46-4671-b633-bf087e892e26",
		Name:          "Lorehold Campus",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("R", "W")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
