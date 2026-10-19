package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fly — Enchantment — Aura {U}:
//
//	"Enchant creature
//	 Enchanted creature has flying and "Whenever this creature deals
//	 combat damage to a player, venture into the dungeon.""
//
// The trigger is granted to the creature (ADR 0093), so it is the
// creature's, and the creature's controller ventures — not the Aura's,
// when the two differ.
//
// No simplification.
func init() {
	const grant = "fly/venture"
	Register(Spec{
		OracleID:     "da1b633f-82b5-4f3e-884c-58c2fe4fcb0f",
		Name:         "Fly",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "Whenever this creature deals combat damage to a player, venture into the dungeon.",
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Fly — venture into the dungeon", Venture),
			},
		}},
		Static: []game.StaticAbility{
			GrantToAttached("flying"),
			GrantAbilitiesToAttached(grant),
		},
	})
}
