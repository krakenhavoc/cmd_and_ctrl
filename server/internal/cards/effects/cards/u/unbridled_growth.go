package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unbridled Growth — {G} Enchantment — Aura (Aether Revolt, #5536):
//
//	"Enchant land
//	 Enchanted land has "{T}: Add one mana of any color."
//	 Sacrifice this Aura: Draw a card."
//
// The mana ability is granted to the land (Abundant Growth's grant); the
// sacrifice is the Aura's own activated ability, with no mana or tap.
//
// No simplification.
func init() {
	const grant = "unbridled-growth/any-color"
	effects.Register(effects.Spec{
		OracleID:     "d749d075-6135-486a-a94a-36c3142960d7",
		Name:         "Unbridled Growth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add one mana of any color.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			}},
		}},
		Static: []game.StaticAbility{effects.GrantAbilities(effects.AttachedToSource, grant)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this Aura: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
