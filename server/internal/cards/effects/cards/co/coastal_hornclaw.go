package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coastal Hornclaw — Creature — Bird {4}{U}, 3/3 (Eighth Edition, S82 OLD batch
// 612, #5607):
//
//	"Sacrifice a land: This creature gains flying until end of turn."
//
// Spitting Spider's cost (a land sacrificed as the ability is activated,
// B08SacrificeALand) with a keyword grant to itself until end of turn
// (CR 611.2a); a Hornclaw that has left the battlefield by resolution gains
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27d672a7-f73b-483d-80a3-a4cd2a9855f2",
		Name:         "Coastal Hornclaw",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a land: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.B08SacrificeALand(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"}, Label: "Coastal Hornclaw — flying"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
