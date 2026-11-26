package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Madblind Mountain — Land — Mountain (Shadowmoor, #5591):
//
//	"({T}: Add {R}.)
//	 This land enters tapped.
//	 {R}, {T}: Shuffle your library. Activate only if you control two or
//	 more red permanents."
//
// Leechridden Swamp's shape: the {R} is the Mountain type's intrinsic
// ability, the activation condition is ControlsAtLeast(2, red) (CR 602.1b),
// and the land is colourless so it never counts toward its own two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ee0b090-3f1e-49d6-bcad-91e0cf1d12ae",
		Name:         "Madblind Mountain",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:     "{R}, {T}: Shuffle your library. Activate only if you control two or more red permanents.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Condition: effects.ControlsAtLeast(2, effects.MatchColor("R")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ShuffleLibrary{}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
