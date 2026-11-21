package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abundant Growth — Enchantment — Aura {G}:
//
//	"Enchant land
//	 When this Aura enters, draw a card.
//	 Enchanted land has '{T}: Add one mana of any color.'"
//
// The Aura shape of ADR 0093's grant, same as Paradise Mantle's
// Equipment shape: EnchantLand for the enchant clause (attachments.go),
// AttachedToSource for the grant's AppliesTo, and an ordinary
// WhenThisEnters trigger for the cantrip.
//
// No simplification.
func init() {
	const grant = "abundant-growth/any-color"
	effects.Register(effects.Spec{
		OracleID:     "947a2665-2f4d-4193-8768-118f85334549",
		Name:         "Abundant Growth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add one mana of any color.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			}},
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(effects.AttachedToSource, grant),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Abundant Growth — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
