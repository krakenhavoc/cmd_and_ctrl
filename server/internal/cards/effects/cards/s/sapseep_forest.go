package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sapseep Forest — Land — Forest (Commander 2021):
//
//	"({T}: Add {G}.)
//	 This land enters tapped.
//	 {G}, {T}: You gain 1 life. Activate only if you control two or more green
//	 permanents."
//
// Leechridden Swamp's shape for green: the {G} is the Forest type's
// intrinsic ability, the land enters tapped, and the life gain is gated on
// two or more green permanents, counted when the ability is activated
// (CR 602.5). The land is colourless, so it never counts toward its own two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d4dcab0-86e5-4ff8-a90f-78a062664e16",
		Name:         "Sapseep Forest",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:     "{G}, {T}: You gain 1 life. Activate only if you control two or more green permanents.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Condition: effects.ControlsAtLeast(2, effects.MatchColor("G")),
			Effect:    effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
}
