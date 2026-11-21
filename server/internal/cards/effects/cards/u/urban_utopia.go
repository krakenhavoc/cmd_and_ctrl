package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urban Utopia — Enchantment — Aura {1}{G} (Guilds of Ravnica, #5525):
//
//	"Enchant land
//	 When this Aura enters, draw a card.
//	 Enchanted land has "{T}: Add one mana of any color.""
//
// Abundant Growth's text: the enter trigger draws, and the enchanted land
// is granted an any-colour mana ability while the Aura is attached.
//
// No simplification.
func init() {
	const grant = "urban-utopia/any-color"
	effects.Register(effects.Spec{
		OracleID:     "77b73f11-a59a-4f2b-83d3-e6fe5064bc87",
		Name:         "Urban Utopia",
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
			effects.WhenThisEnters("Urban Utopia — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
