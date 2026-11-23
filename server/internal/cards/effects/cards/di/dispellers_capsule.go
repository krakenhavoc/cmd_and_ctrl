package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dispeller's Capsule — Artifact {W} (Shards of Alara, #5588):
//
//	"{2}{W}, {T}, Sacrifice this artifact: Destroy target artifact or
//	 enchantment."
//
// Executioner's Capsule's row with Naturalize's target. The tap and the
// sacrifice are costs, so the Capsule is gone before anyone can respond
// (CR 602.2b, CR 701.21a), and a target gone in response is untouched
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91fc862d-fdde-45ea-a05d-30d38e7a735c",
		Name:         "Dispeller's Capsule",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}, Sacrifice this artifact: Destroy target artifact or enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{2}{W}"},
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
