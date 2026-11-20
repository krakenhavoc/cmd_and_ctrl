package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Freed from the Real — Enchantment — Aura {2}{U}:
//
//	"Enchant creature
//	 {U}: Tap enchanted creature.
//	 {U}: Untap enchanted creature."
//
// Both abilities belong to the Aura, so its controller activates
// them whoever controls the creature, and neither targets: the
// creature is found through the attachment as the ability resolves.
// An Aura that has fallen off by then taps nothing (the ability still
// resolves, CR 608.2b does not apply to an untargeted effect).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03d32ace-dd99-48c6-bcae-9b915d041aef",
		Name:         "Freed from the Real",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}: Tap enchanted creature",
				Purpose: game.Purpose{Answers: game.AnswerRestrict},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.TapAttachedHost,
			},
			{
				Label:   "{U}: Untap enchanted creature",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.UntapAttachedHost,
			},
		},
	})
}
