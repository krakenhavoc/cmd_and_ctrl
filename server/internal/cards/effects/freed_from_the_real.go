package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "03d32ace-dd99-48c6-bcae-9b915d041aef",
		Name:         "Freed from the Real",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Activated: []ActivatedAbility{
			{
				Label:   "{U}: Tap enchanted creature",
				Purpose: game.Purpose{Answers: game.AnswerRestrict},
				Cost:    ManaCost("{U}"),
				Effect:  tapAttachedHost,
			},
			{
				Label:   "{U}: Untap enchanted creature",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{U}"),
				Effect:  untapAttachedHost,
			},
		},
	})
}
