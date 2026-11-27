package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Arsonist — Creature — Human Soldier {2}{R}, 1/1 (Prophecy, S82 OLD b672,
// #5617):
//
//	"{1}, Sacrifice two lands: Destroy target land."
//
// Barbarian Riftcutter's destroy-target-land with a mana cost and a two-land
// sacrifice (SacrificeN). The lands are paid at announce; a target land that was
// sacrificed to pay, or that left in response, is not destroyed (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d32f973e-38db-4f74-aec6-258d25c1fb7d",
		Name:         "Keldon Arsonist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice two lands: Destroy target land.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(2, "two lands", effects.Land())),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
