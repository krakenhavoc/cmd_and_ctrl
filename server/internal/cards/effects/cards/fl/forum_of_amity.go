package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forum of Amity — Land (Secrets of Strixhaven, #4456):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}.
//	 {2}{W}{B}, {T}: Surveil 1."
//
// Silverquill Campus's shape with a surveil: a nonbasic land with no basic
// types, so both mana colours are declared, and the surveil is an ordinary
// activated ability whose cost taps the land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "349ea6c7-6b3e-417f-b082-b712e2b1635b",
		Name:          "Forum of Amity",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("W", "B")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{B}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}{B}"), effects.TapCost()),
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}
