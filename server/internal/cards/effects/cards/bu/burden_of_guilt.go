package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burden of Guilt — Enchantment — Aura {W} (The List):
//
//	"Enchant creature
//	 {1}: Tap enchanted creature."
//
// Freed from the Real's tap half: the ability belongs to the Aura, so its
// controller activates it whoever controls the creature, and it does not
// target (the creature is found through the attachment as it resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ae88ba6-4973-45d1-a251-4d924ab4bc73",
		Name:         "Burden of Guilt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Tap enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.ManaCost("{1}"),
			Effect:  effects.TapAttachedHost,
		}},
	})
}
