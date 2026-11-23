package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diamond Pick-Axe — Artifact — Equipment {R} (The Lost Caverns of
// Ixalan, #4470):
//
//	"Indestructible
//	 Equipped creature gets +1/+1 and has "Whenever this creature attacks,
//	 create a Treasure token."
//	 Equip {2}"
//
// The attack trigger is an ability the equipped creature has (ADR 0093,
// Iconic Shield's shape) with Rocketeer Boostbuggy's Treasure body, so it
// belongs to the creature and its controller makes the Treasure.
// Indestructible is the Equipment's own printed keyword.
//
// No simplification.
func init() {
	const grantKey = "diamond-pick-axe/attack-treasure"
	effects.Register(effects.Spec{
		OracleID:        "542e7d60-3fc8-46e2-8786-f7c8b3d4c979",
		Name:            "Diamond Pick-Axe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("Diamond Pick-Axe — create a Treasure token",
					effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
			},
			Text: "Whenever this creature attacks, create a Treasure token.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(grantKey),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
