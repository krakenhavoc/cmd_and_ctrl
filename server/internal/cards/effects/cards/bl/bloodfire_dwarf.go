package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodfire Dwarf — Creature — Dwarf {R}, 1/1 (Apocalypse, #5614):
//
//	"{R}, Sacrifice this creature: It deals 1 damage to each creature
//	 without flying."
//
// Bloodfire Kavu's shape; flying is read through the layered view as the
// ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff42551a-a08e-4d0c-a5b1-d1c1bbff4915",
		Name:         "Bloodfire Dwarf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: It deals 1 damage to each creature without flying.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 1)
			},
		}},
	})
}
