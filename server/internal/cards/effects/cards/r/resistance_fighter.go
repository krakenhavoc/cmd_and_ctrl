package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resistance Fighter — Creature — Human Soldier {W}:
//
//	"Sacrifice this creature: Prevent all combat damage target creature would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the ability resolves (CR 400.7).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a95a2d1c-70a9-4eb2-ae30-a07fab26a9c6",
		Name:         "Resistance Fighter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.ShieldAgainstTargetsRow(
			"Sacrifice this creature: Prevent all combat damage target creature would deal this turn.",
			effects.SacrificeThis(), effects.TargetCreature("target creature"), true)},
	})
}
