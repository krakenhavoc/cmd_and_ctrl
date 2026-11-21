package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armillary Sphere — Artifact {2} (Conflux, #5587):
//
//	"{2}, {T}, Sacrifice this artifact: Search your library for up to two
//	 basic land cards, reveal them, put them into your hand, then
//	 shuffle."
//
// Journey of Discovery's search on a sacrifice row. "Up to two" lets the
// searcher take none (CR 701.23d), and the artifact is sacrificed as a
// cost, so it is gone before the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3963140c-da67-43e6-9514-fe9dc0a43c4d",
		Name:         "Armillary Sphere",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Search your library for up to two basic land cards, reveal them, put them into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     2,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Armillary Sphere — up to two basic land cards",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
