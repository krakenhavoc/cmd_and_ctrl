package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "542e7d60-3fc8-46e2-8786-f7c8b3d4c979",
		Name:            "Diamond Pick-Axe",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Grants: []AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Diamond Pick-Axe — create a Treasure token",
					Do(CreateToken{Template: TreasureToken(), N: 1})),
			},
			Text: "Whenever this creature attacks, create a Treasure token.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(grantKey),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
