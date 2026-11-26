package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const crystallineNautilusGrant = "crystalline-nautilus/sacrifice-when-targeted"

// Crystalline Nautilus — Enchantment Creature — Nautilus {2}{U}, 4/4 (Journey into Nyx):
//
//	"Bestow {3}{U}{U}
//	 When this permanent becomes the target of a spell or ability, sacrifice it.
//	 Enchanted creature gets +4/+4 and has "When this creature becomes the
//	 target of a spell or ability, sacrifice it.""
//
// The Nautilus's own trigger is the permanent's, creature or bestowed Aura
// (Skulking Ghost's shape). The host's is granted, so it sacrifices the host.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77661282-1c98-4f70-b617-c75633ab04c8",
		Name:         "Crystalline Nautilus",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{3}{U}{U}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self, "Crystalline Nautilus — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
		Grants: []effects.AbilityGrant{{
			Key: crystallineNautilusGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventBecomesTarget, effects.Self, "Crystalline Nautilus — sacrifice the enchanted creature", effects.SacrificeThisIfStillOnBattlefield),
			},
			Text: "When this creature becomes the target of a spell or ability, sacrifice it.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(4, 4),
			effects.GrantAbilitiesToAttached(crystallineNautilusGrant),
		},
	})
}
