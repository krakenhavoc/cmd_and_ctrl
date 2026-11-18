package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "894176e3-af9c-41ef-9f35-59dfdda3c17f",
		Name:         "Spark Reaper",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
		Activated: []ActivatedAbility{{
			Label:   "{3}, Sacrifice a creature or planeswalker: You gain 1 life and draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{3}"), SacrificeN(1, "a creature or planeswalker", Or(Creature(), Planeswalker()))),
			Effect:  Do(GainLife{Amount: 1}, DrawCards{N: 1}),
		}},
	})
}
