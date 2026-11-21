package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weirding Wood — Enchantment — Aura {2}{G}:
//
//	"Enchant land
//	 When this Aura enters, investigate. (Create a Clue token. It's an
//	 artifact with "{2}, Sacrifice this token: Draw a card.")
//	 Enchanted land has "{T}: Add two mana of any one color.""
//
// Grafted Growth's granted mana bundle (ADR 0093) on the enchanted land,
// with an enter trigger that investigates for the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ce5cc2f-e223-4c3b-adc1-c0109fd31762",
		Name:         "Weirding Wood",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 1},
		Grants: []effects.AbilityGrant{{
			Key:  weirdingWoodMana,
			Text: "{T}: Add two mana of any one color.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: effects.OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(weirdingWoodMana)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Weirding Wood — investigate", effects.Do(effects.Investigate{N: 1})),
		},
	})
}

const weirdingWoodMana = "weirding-wood/mana"
