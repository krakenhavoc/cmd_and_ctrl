package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "4c76db22-1992-448d-be4a-a9095420b8cc",
		Name:            "Silkwing Scout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{G}, Sacrifice this creature: Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{G}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Silkwing Scout — a basic land",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
