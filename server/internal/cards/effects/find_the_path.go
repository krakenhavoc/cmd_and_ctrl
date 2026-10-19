package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "860eb9eb-c318-4eae-9b54-088f667115ec",
		Name:         "Find the Path",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add {G}{G}.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}{G}",
				Label:    "Add {G}{G}",
			}},
		}},
		Static:    []game.StaticAbility{GrantAbilities(AttachedToSource, grant)},
		Triggered: []game.TriggeredAbility{WhenThisEntersVenture("Find the Path")},
	})
}
