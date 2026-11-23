package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// River Boa — Creature — Snake {1}{G}, 2/1 (Visions, #5631):
//
//	"Islandwalk (This creature can't be blocked as long as defending player
//	 controls an Island.)
//	 {G}: Regenerate this creature."
//
// Islandwalk is landwalk, an engine keyword carried as printed data
// (CR 702.14c). The ability makes a regeneration shield: the next time
// the Boa would be destroyed this turn, its damage is removed and it is
// tapped and removed from combat instead (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86187aed-77ce-4b2d-aab0-0a8f807a9451",
		Name:            "River Boa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
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
