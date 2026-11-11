package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twisted Abomination — Creature {5}{B}, 5/3:
//
//	"{B}: Regenerate this creature.
//	 Swampcycling {2}"
func init() {
	Register(Spec{
		OracleID:     "139af10d-a3e5-4ade-b876-87324e6c8017",
		Name:         "Twisted Abomination",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{B}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
			Typecycling("Swampcycling", "{2}", "a Swamp card", IsLandWithSubtype("Swamp")),
		},
	})
}
