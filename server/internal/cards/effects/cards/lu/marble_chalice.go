package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marble Chalice — Artifact {2}{W} (Shards of Alara, S82 OLD batch 512,
// #5588):
//
//	"{T}: You gain 1 life."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "236f318e-9bd1-4eb0-8b19-af015b3a74ea",
		Name:         "Marble Chalice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You gain 1 life.",
			Purpose: game.Purpose{LifeGain: 1, Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
