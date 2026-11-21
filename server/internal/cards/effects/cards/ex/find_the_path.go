package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Find the Path — Enchantment — Aura {2}{G}:
//
//	"Enchant land
//	 When this Aura enters, venture into the dungeon.
//	 Enchanted land has "{T}: Add {G}{G}.""
//
// The mana ability is granted to the enchanted land (Abundant Growth's
// shape), so it is the land's and its controller's.
//
// No simplification.
func init() {
	const grant = "find-the-path/gg"
	effects.Register(effects.Spec{
		OracleID:     "860eb9eb-c318-4eae-9b54-088f667115ec",
		Name:         "Find the Path",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add {G}{G}.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{G}{G}",
				Label:    "Add {G}{G}",
			}},
		}},
		Static:    []game.StaticAbility{effects.GrantAbilities(effects.AttachedToSource, grant)},
		Triggered: []game.TriggeredAbility{effects.WhenThisEntersVenture("Find the Path")},
	})
}
