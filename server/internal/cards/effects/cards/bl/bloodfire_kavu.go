package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodfire Kavu — Creature — Kavu {2}{R}{R}, 2/2 (Apocalypse, #5614):
//
//	"{R}, Sacrifice this creature: It deals 2 damage to each creature."
//
// The sacrifice is a cost (CR 602.2b), so the damage comes from the Kavu as
// it last existed (CR 113.7a); it is already gone, so it hits every other
// creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a12880c4-1987-42af-8595-0b88981e1091",
		Name:         "Bloodfire Kavu",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: It deals 2 damage to each creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.Creature(), 2)
			},
		}},
	})
}
