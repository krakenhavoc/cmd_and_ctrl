package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Pine Needles — Creature — Plant Wall {3}{G}, 3/3 (Ice Age,
// #5636):
//
//	"Defender (This creature can't attack.)
//	 {G}: Regenerate this creature."
//
// Defender is a printed keyword (CR 702.3b). The regeneration row is
// Drowned's with a {G} cost: Regenerate on the ability's own source, a
// shield that replaces its next destruction this turn (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d5bc31c-f4a1-4a59-b466-0347ddbbc7fb",
		Name:            "Wall of Pine Needles",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
