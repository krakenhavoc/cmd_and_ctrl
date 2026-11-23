package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Cratermaker — Creature — Goblin Warrior {1}{R} (The Hobbit Commander), 2/2:
//
//	"{1}, Sacrifice this creature: Choose one —
//	 • This creature deals 2 damage to target creature.
//	 • Destroy target colorless nonland permanent."
//
// #4454, S76. A modal activated ability is one ability per bullet with
// the same cost (Insidious Fungus's shape): the mode is chosen at
// activation (CR 700.2) and the client's ability menu is the picker.
// The damage is dealt by the Cratermaker as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	cost := effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis())
	effects.Register(effects.Spec{
		OracleID:     "1962d9b3-f057-4d2c-9c5b-386d9a66148a",
		Name:         "Goblin Cratermaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, Sacrifice this creature: This creature deals 2 damage to target creature.",
				Cost:    cost,
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
			},
			{
				Label:   "{1}, Sacrifice this creature: Destroy target colorless nonland permanent.",
				Cost:    cost,
				Targets: effects.TargetPermanent("target colorless nonland permanent", effects.And(effects.Colorless(), effects.Nonland())),
				Effect:  effects.DestroyFirstLegalTarget,
			},
		},
	})
}
