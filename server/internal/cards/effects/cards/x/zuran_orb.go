package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zuran Orb — Artifact, {0} (EDHREC rank 910):
//
//	"Sacrifice a land: You gain 2 life."
//
// The free sac outlet for lands — a Gitrog or Titania deck's engine
// piece, and a no-mana answer to a board wipe you can see coming. One
// activated ability whose cost is "sacrifice a land", built through
// the same clause the creature sac outlets use, so the client opens
// the same picker and the engine filters it to the activator's own
// lands (CR 701.21a). No tap, so it can be activated any number of
// times in a row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08cb8a30-9cb4-4517-bee5-8848aa60d1a2",
		Name:         "Zuran Orb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a land: You gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.B08SacrificeALand(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
