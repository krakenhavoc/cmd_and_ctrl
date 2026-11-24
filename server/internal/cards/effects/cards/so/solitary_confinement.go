package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Solitary Confinement — Enchantment {2}{W} (Modern Horizons 2, #5501):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless you
//	 discard a card.
//	 Skip your draw step.
//	 You have shroud.
//	 Prevent all damage that would be dealt to you."
//
// The upkeep trigger asks for a discard of up to one card and sacrifices the
// enchantment on a declined or impossible one (Argentum Masticore's shape,
// only while it is still the permanent that triggered). The draw step is
// skipped (Yawgmoth's Bargain's replacement), shroud is the player keyword
// of Spirit of the Hearth's kind (CR 702.18b), and all damage to its
// controller is prevented, combat and noncombat alike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "d41ab41d-07c8-4f6d-be5e-7aec9f5f6365",
		Name:           "Solitary Confinement",
		Completeness:   effects.CompletenessFull,
		PlayerKeywords: []string{effects.KeywordShroud},
		Replacements: []game.ReplacementEffect{
			effects.SkipYourDrawStep(),
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToYou,
				Damage: effects.AnyDamage,
				Label:  "Solitary Confinement — prevent all damage that would be dealt to you",
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Solitary Confinement — sacrifice this enchantment unless you discard a card",
				effects.OldB030SacrificeUnlessDiscard("Solitary Confinement")),
		},
	})
}
