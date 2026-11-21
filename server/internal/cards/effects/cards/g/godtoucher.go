package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Godtoucher — Creature — Elf Cleric {3}{G}, 2/2:
//
//	"{1}{W}, {T}: Prevent all damage that would be dealt to target
//	 creature with power 5 or greater this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the not-one-use shield pinned to
// the target. The power is checked as the target is chosen and again as
// the ability resolves (CR 608.2b); the shield then lasts the turn,
// whatever the creature's power becomes.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3094f214-5aa8-4003-8a1e-035485dfb978",
		Name:         "Godtoucher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.ShieldTargetRow(
			"{1}{W}, {T}: Prevent all damage that would be dealt to target creature with power 5 or greater this turn.",
			effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			effects.TargetCreature("target creature with power 5 or greater", effects.PowerGE(5)))},
	})
}
