package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const trollhideRegenerateGrant = "trollhide/regenerate"

// Trollhide — Enchantment — Aura {2}{G} (Magic 2012, #5573):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has "{1}{G}: Regenerate this
//	 creature." (The next time the creature would be destroyed this
//	 turn, instead tap it, remove it from combat, and heal all damage
//	 on it.)"
//
// Dragon Mantle's shape: PumpAttached, and the regeneration row granted to
// the enchanted creature (GrantAbilitiesToAttached), so the creature's
// controller activates it and the shield goes on the creature (Albino
// Troll's row).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "397df813-c759-4246-8def-122c7ff82d88",
		Name:         "Trollhide",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: trollhideRegenerateGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}{G}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{1}{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			}},
			Text: "{1}{G}: Regenerate this creature.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantAbilitiesToAttached(trollhideRegenerateGrant),
		},
	})
}
