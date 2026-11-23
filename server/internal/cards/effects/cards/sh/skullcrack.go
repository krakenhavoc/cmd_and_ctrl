package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullcrack — Instant {1}{R}:
//
//	"Players can't gain life this turn. Damage can't be prevented this
//	 turn. Skullcrack deals 3 damage to target player or planeswalker."
//
// ADR 0107 §5 (#1853, #1880): both rule grants are stored records
// (ModCantGainLife, ModDamageCantBePrevented) that end at cleanup
// (CR 514.2), and both begin before the damage, so the 3 damage
// itself can't be prevented. "Players" is every player, the caster
// included.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6fc0b0c-9ba2-47f2-ab92-26fd20d0f86d",
		Name:         "Skullcrack",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		OnResolve:    effects.SkullcrackShape(3),
	})
}
