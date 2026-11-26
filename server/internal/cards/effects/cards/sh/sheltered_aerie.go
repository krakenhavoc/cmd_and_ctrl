package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sheltered Aerie — Enchantment — Aura {2}{G} (Dragons of Tarkir, S82 old batch
// 332):
//
//	"Enchant land
//	 Enchanted land has "{T}: Add two mana of any one color.""
//
// Weirding Wood's granted mana bundle (ADR 0093) on the enchanted land, without the
// investigate: the land's controller taps it for the mana, two of one colour of
// their choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2c408f6-3994-4a90-a9f5-f248eb976017",
		Name:         "Sheltered Aerie",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key:  shelteredAerieMana,
			Text: "{T}: Add two mana of any one color.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: effects.OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(shelteredAerieMana)},
	})
}

// shelteredAerieMana names the granted mana ability's bundle: an on-disk identity.
const shelteredAerieMana = "sheltered-aerie/mana"
