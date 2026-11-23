package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paradox Gardens — Land (Secrets of Strixhaven, #4456):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}.
//	 {2}{G}{U}, {T}: Surveil 1."
//
// University Campus's shape with a different price: a nonbasic land with no
// basic types, so both mana colours are declared, and an ordinary activated
// ability whose cost taps the land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "638ff242-63d5-457d-a7a6-40ad51052e2e",
		Name:          "Paradox Gardens",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("G", "U")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{U}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}{U}"), effects.TapCost()),
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}
