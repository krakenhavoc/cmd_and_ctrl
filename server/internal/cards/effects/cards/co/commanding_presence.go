package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commanding Presence — Enchantment — Aura {3}{W}:
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has first strike and "Whenever this
//	 creature deals combat damage to a player, create a 1/1 white Human
//	 Soldier creature token.""
//
// The granted trigger belongs to the enchanted creature, so its controller
// makes the token and it stops with the Aura.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03160a74-8b83-4e07-b702-977a91a6311f",
		Name:         "Commanding Presence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 1},
		Grants: []effects.AbilityGrant{{
			Key: commandingPresenceGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Commanding Presence — create a 1/1 white Human Soldier creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, create a 1/1 white Human Soldier creature token.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("first strike"),
			effects.GrantAbilitiesToAttached(commandingPresenceGrant),
		},
	})
}

// commandingPresenceGrant is the granted bundle's key, an on-disk identity:
// never renamed, never reused.
const commandingPresenceGrant = "commanding-presence/human-soldier"
