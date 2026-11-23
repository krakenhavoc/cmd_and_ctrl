package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yavimaya Gnats — Creature — Insect {2}{G}, 0/1 (Ice Age, #5636):
//
//	"Flying
//	 {G}: Regenerate this creature."
//
// Flying is a printed keyword (CR 702.9b). The regeneration row is
// Drowned's with a {G} cost: Regenerate on the ability's own source, a
// shield that replaces its next destruction this turn (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "828c1a94-36e1-484b-afc8-87f85956d60d",
		Name:            "Yavimaya Gnats",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
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
