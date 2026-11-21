package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tangle Hulk — Artifact Creature — Phyrexian Beast {5}, 5/3 (Mirrodin Besieged,
// #5576):
//
//	"{2}{G}: Regenerate this creature."
//
// Undercity Troll's regeneration: a shield on the ability's own source
// (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6d0360c-ebc6-48b5-80ad-d52dadcdd5f5",
		Name:         "Tangle Hulk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
