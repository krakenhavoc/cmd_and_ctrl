package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jungle Troll — Creature — Troll {1}{R}{G}, 2/1 (Mirage, #5633):
//
//	"{R}: Regenerate this creature.
//	 {G}: Regenerate this creature."
//
// Two of Restless Dead's rows, one per colour: Regenerate on the
// ability's own source, a shield that replaces the next destruction this
// turn (CR 701.19a). Each activation is its own shield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1bb17d7-5396-4c99-88bb-f7ab7d47b023",
		Name:         "Jungle Troll",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{R}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{G}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
