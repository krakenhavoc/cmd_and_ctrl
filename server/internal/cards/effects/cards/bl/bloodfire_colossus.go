package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodfire Colossus — Creature — Giant {6}{R}{R}, 6/6 (Apocalypse, #5614):
//
//	"{R}, Sacrifice this creature: It deals 6 damage to each creature and
//	 each player."
//
// One damage event over every creature and every player, the activator
// included (Psychotic Haze's body at 6), dealt by the Colossus as it last
// existed (CR 113.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc59e28a-eda3-468c-a31b-73e4b615f953",
		Name:         "Bloodfire Colossus",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: It deals 6 damage to each creature and each player.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6}},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 6)
			},
		}},
	})
}
