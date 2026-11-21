package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spark Reaper — Creature — Zombie {2}{B}, 2/3:
//
//	"{3}, Sacrifice a creature or planeswalker: You gain 1 life and draw
//	 a card."
//
// The sacrifice is a cost (the Spark Reaper itself qualifies). The life
// and the card are one resolution, in the printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "894176e3-af9c-41ef-9f35-59dfdda3c17f",
		Name:         "Spark Reaper",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, Sacrifice a creature or planeswalker: You gain 1 life and draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.SacrificeN(1, "a creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
			Effect:  effects.Do(effects.GainLife{Amount: 1}, effects.DrawCards{N: 1}),
		}},
	})
}
