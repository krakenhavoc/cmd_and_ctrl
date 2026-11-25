package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molten Nursery — Enchantment {2}{R} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Whenever you cast a colorless spell, this enchantment deals 1 damage to any
//	 target."
//
// A colorless-spell cast trigger with a target chosen as it goes on the stack
// (CR 603.3d) and re-checked as it resolves (CR 608.2b); the damage is dealt by
// the enchantment, last known if it has gone (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "92abb79d-3c60-48f7-87a0-13c25a5e1128",
		Name:            "Molten Nursery",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WheneverYouCast(effects.Colorless(),
				"Molten Nursery — this enchantment deals 1 damage to any target",
				effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
