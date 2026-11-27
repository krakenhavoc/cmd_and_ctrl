package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandsower — Creature — Spirit {3}{W}, 1/3 (Ravnica: City of Guilds, S82 OLD
// batch 574, #5600):
//
//	"Tap three untapped creatures you control: Tap target creature."
//
// The cost is a tap-others component, not the {T} symbol (Siege Zombie's),
// so Sandsower may tap itself as one of the three and a creature that arrived
// this turn may pay (CR 302.6). The target is chosen first and is tapped as
// the ability resolves; one that is gone is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1dce1f39-eacb-4d29-899b-55569284c7ed",
		Name:         "Sandsower",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap three untapped creatures you control: Tap target creature.",
			Cost:    effects.S82MidB09TapThreeCreatures(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
