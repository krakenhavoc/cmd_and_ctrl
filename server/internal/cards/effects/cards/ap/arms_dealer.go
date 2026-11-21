package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arms Dealer — Creature — Goblin Rogue {2}{R}, 1/1 (Jumpstart 2022,
// #5948):
//
//	"{1}{R}, Sacrifice a Goblin: This creature deals 4 damage to target
//	 creature."
//
// Siege-Gang Lieutenant's cost: any Goblin its controller controls may be
// sacrificed, the Arms Dealer itself included, and the damage is then dealt
// by it as it last existed (CR 608.2h). The target is chosen as the ability
// is activated, so the sacrificed Goblin is not a legal pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ebc4198-7317-4ffb-b8b8-14733c2077ff",
		Name:         "Arms Dealer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a Goblin: This creature deals 4 damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), game.AbilityCost{SacrificeOther: effects.B12SacrificeAGoblin()}),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
			Effect:  effects.DamageToFirstTarget(4),
		}},
	})
}
