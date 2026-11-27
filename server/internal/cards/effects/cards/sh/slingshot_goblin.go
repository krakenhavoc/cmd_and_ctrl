package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slingshot Goblin — Creature — Goblin {2}{R}, 2/2 (Planeshift, S82 OLD
// #5615):
//
//	"{R}, {T}: This creature deals 2 damage to target blue creature."
//
// One target clause (a blue creature, judged at announce and on resolution,
// CR 608.2b) and the shared "deals N damage to the target" body, the source
// being the Goblin. The {T} keeps it from being activated the turn it arrives
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b377595f-c6c2-4c54-b462-73ef4eba519b",
		Name:         "Slingshot Goblin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: This creature deals 2 damage to target blue creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    game.AbilityCost{Mana: "{R}", Tap: true},
			Targets: effects.TargetCreature("target blue creature", effects.OfColor("U")),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
