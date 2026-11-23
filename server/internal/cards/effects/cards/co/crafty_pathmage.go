package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crafty Pathmage — Creature — Human Wizard {2}{U}, 1/1 (Onslaught):
//
//	"{T}: Target creature with power 2 or less can't be blocked this
//	 turn."
//
// Dwarven Warriors' ability. The power limit is checked at announce and
// again at resolution (CR 608.2b), so a creature pumped past 2 in
// response is no longer a legal target and is not made unblockable.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4020178-71d9-4363-a0bd-b9a7eecd5d02",
		Name:         "Crafty Pathmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature with power 2 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
			Effect:  effects.TargetCantBeBlockedThisTurn("Crafty Pathmage"),
		}},
	})
}
