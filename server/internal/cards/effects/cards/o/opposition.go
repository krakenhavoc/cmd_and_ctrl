package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Opposition — Enchantment {2}{U}{U} (Urza's Destiny, #5621):
//
//	"Tap an untapped creature you control: Tap target artifact,
//	 creature, or land."
//
// The cost is tapping a creature, not the {T} symbol (CR 302.6), so a
// creature that arrived this turn can pay it, once per untapped
// creature. The target is any player's artifact, creature or land; one
// gone by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf0b252d-1295-44e4-bac3-113b5732a2a9",
		Name:         "Opposition",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: Tap target artifact, creature, or land.",
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Targets: effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
