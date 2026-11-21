package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soliton — Artifact Creature — Construct {5}, 3/4 (Scars of Mirrodin, #5577):
//
//	"{U}: Untap this creature."
//
// Blizzard Elemental's untap at {U}; an untapped Soliton may still
// activate it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c04b4323-db6a-4ab1-99b7-eef3acad3c4f",
		Name:         "Soliton",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "{U}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
