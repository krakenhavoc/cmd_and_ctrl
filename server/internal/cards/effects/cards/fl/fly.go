package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "da1b633f-82b5-4f3e-884c-58c2fe4fcb0f",
		Name:         "Fly",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "Whenever this creature deals combat damage to a player, venture into the dungeon.",
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Fly — venture into the dungeon", effects.Venture),
			},
		}},
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying"),
			effects.GrantAbilitiesToAttached(grant),
		},
	})
}
