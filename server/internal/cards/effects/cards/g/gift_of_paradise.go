package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gift of Paradise — Enchantment — Aura {2}{G} (Commander 2021, S82 OLD b058,
// #5504):
//
//	"Enchant land
//	 When this Aura enters, you gain 3 life.
//	 Enchanted land has '{T}: Add two mana of any one color.'"
//
// Weirding Wood's granted mana bundle on the enchanted land, with a
// gain-3-life enter trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43a86d36-9653-4498-a533-8b30b919cdbf",
		Name:         "Gift of Paradise",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key:  giftOfParadiseMana,
			Text: "{T}: Add two mana of any one color.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: effects.OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(giftOfParadiseMana)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gift of Paradise — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}

const giftOfParadiseMana = "gift-of-paradise/mana"
