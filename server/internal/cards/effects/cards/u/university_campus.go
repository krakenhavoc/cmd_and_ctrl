package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// University Campus — Land (Marvel's Spider-Man, #4460):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}.
//	 {4}, {T}: Surveil 1."
//
// A nonbasic land with no basic types, so both mana colours are declared. The
// surveil is an ordinary activated ability whose cost taps the land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "98e547de-b963-4ee4-9a08-67bae010734b",
		Name:          "University Campus",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("W", "U")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}
