package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leechridden Swamp — Land — Swamp (EDHREC rank 2502):
//
//	"({T}: Add {B}.)
//	 This land enters tapped.
//	 {B}, {T}: Each opponent loses 1 life. Activate only if you control
//	 two or more black permanents."
//
// Shadowmoor's black hybrid-land drain. The {B} is the Swamp type's
// intrinsic ability, which the engine derives from the type line, so
// the spec declares none. "Activate only if you control two or more
// black permanents" is the drain's activation condition (CR 602.1b,
// #743), ControlsAtLeast(2, black) over effective colours. The land
// itself is colourless, so it never counts toward its own two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d83c86c1-126d-49e9-9b13-9e55784c49c5",
		Name:         "Leechridden Swamp",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:     "{B}, {T}: Each opponent loses 1 life. Activate only if you control two or more black permanents.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			Condition: effects.ControlsAtLeast(2, effects.MatchColor("B")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.EachOpponentLosesLifeNow(g, item, 1)
			},
		}},
	})
}
