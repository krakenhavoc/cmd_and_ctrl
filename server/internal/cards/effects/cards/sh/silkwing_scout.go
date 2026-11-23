package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silkwing Scout — Creature — Faerie Scout {2}{U}, 2/1 (Dissension, #5598):
//
//	"Flying
//	 {G}, Sacrifice this creature: Search your library for a basic land card,
//	 put that card onto the battlefield tapped, then shuffle."
//
// Sakura-Tribe Elder's search behind a {G}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c76db22-1992-448d-be4a-a9095420b8cc",
		Name:            "Silkwing Scout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice this creature: Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Silkwing Scout — a basic land",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
