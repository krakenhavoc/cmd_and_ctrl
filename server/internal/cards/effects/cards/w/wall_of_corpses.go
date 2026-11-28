package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Corpses — Creature — Wall {1}{B}, 0/2 (Mirage):
//
//	"Defender
//	 {B}, Sacrifice this creature: Destroy target creature this creature is
//	 blocking."
//
// Tinder Wall's "target creature it's blocking" (BlockedBySource), judged at
// announce and again at resolution, with the Wall sacrificed as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "16a7895a-ca23-4112-a91f-4b1528e75cea",
		Name:            "Wall of Corpses",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice this creature: Destroy target creature this creature is blocking.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeThis()),
			Targets: effects.BlockedBySource(effects.TargetCreature("target creature this creature is blocking")),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
