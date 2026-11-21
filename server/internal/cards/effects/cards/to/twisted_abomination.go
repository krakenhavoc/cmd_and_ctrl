package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twisted Abomination — Creature {5}{B}, 5/3:
//
//	"{B}: Regenerate this creature.
//	 Swampcycling {2}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "139af10d-a3e5-4ade-b876-87324e6c8017",
		Name:         "Twisted Abomination",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			},
			effects.Typecycling("Swampcycling", "{2}", "a Swamp card", effects.IsLandWithSubtype("Swamp")),
		},
	})
}
