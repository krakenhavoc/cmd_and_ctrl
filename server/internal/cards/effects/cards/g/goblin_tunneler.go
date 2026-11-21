package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Tunneler — Creature — Goblin Rogue {1}{R}, 1/1 (Rise of the
// Eldrazi, #5581):
//
//	"{T}: Target creature with power 2 or less can't be blocked this turn."
//
// Crafty Pathmage's ability. The power is checked as the target is chosen
// and again as the ability resolves; a creature pumped past 2 afterwards
// stays unblockable.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1005bc16-8a6a-476a-a647-4ed40d2e96de",
		Name:         "Goblin Tunneler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature with power 2 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
			Effect:  effects.TargetCantBeBlockedThisTurn("Goblin Tunneler"),
		}},
	})
}
