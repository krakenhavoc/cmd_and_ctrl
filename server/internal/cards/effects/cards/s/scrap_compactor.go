package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrap Compactor — Artifact {1} (Aetherdrift, #4464):
//
//	"{3}, {T}, Sacrifice this artifact: It deals 3 damage to target
//	 creature.
//	 {6}, {T}, Sacrifice this artifact: Destroy target creature or
//	 Vehicle."
//
// Bear Trap's sacrifice-for-damage row (the sacrificed Compactor deals
// the damage on its last-known information, CR 113.7a) beside Cathar
// Commando's sacrifice-and-destroy row with Spin Out's "creature or
// Vehicle" target, so an uncrewed Vehicle is a legal pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f113875-a411-463b-8b4d-91d1dee95c2a",
		Name:         "Scrap Compactor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}, {T}, Sacrifice this artifact: It deals 3 damage to target creature.",
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
			},
			{
				Label:   "{6}, {T}, Sacrifice this artifact: Destroy target creature or Vehicle.",
				Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetPermanent("target creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
				Effect:  effects.DestroyChosenPermanent,
			},
		},
	})
}
