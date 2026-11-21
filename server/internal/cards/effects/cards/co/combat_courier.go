package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Combat Courier — Artifact Creature — Construct {1}, 1/1:
//
//	"{2}, Sacrifice this creature: Draw a card.
//	 Unearth {U}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54d32245-c796-4e8e-b61e-c218e3902565",
		Name:         "Combat Courier",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, Sacrifice this creature: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect:  effects.EbDrawACard,
			},
			effects.Unearth("{U}"),
		},
	})
}
