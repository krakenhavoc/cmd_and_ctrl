package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gorm the Great — Legendary Creature — Giant Warrior {3}{G}, 2/7 (Battlebond, #5529):
//
//	"Partner with Virtus the Veiled (When this creature enters, target player may put Virtus into their hand from their library, then shuffle.)"
//	"Vigilance"
//	"Gorm must be blocked if able, and Gorm must be blocked by two or more creatures if able."
//
// effects.MustBeBlockedByTwoOrMore() is the sentence as its two block requirements (#6243).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7195e9f-53fa-4fb5-b1f1-099c0a772090",
		Name:            "Gorm the Great",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          effects.MustBeBlockedByTwoOrMore(),
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Gorm the Great", "Virtus the Veiled"),
		},
	})
}
