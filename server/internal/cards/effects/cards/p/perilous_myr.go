package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perilous Myr — Artifact Creature — Phyrexian Myr {2}, 1/1 (Scars of
// Mirrodin, #5577):
//
//	"When this creature dies, it deals 2 damage to any target."
//
// Shady Informant's dies trigger. The Myr is the source of the damage as
// it last existed on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1e8f746-1837-4ea8-9a07-e57946883353",
		Name:         "Perilous Myr",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Perilous Myr — 2 damage to any target",
				effects.SourceDealsDamageToEachLegalTarget(2)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
