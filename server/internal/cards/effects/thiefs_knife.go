package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thief's Knife — Artifact — Equipment {2}{U} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/+1, has "Whenever this creature deals
//	 combat damage to a player, draw a card," and is a Rogue in addition
//	 to its other types.
//	 Equip {4}"
//
// The combat-damage trigger is a granted ability (ADR 0093), as Black
// Mage's Rod's cast trigger is: it belongs to the equipped creature, goes
// away with the Equipment, and has the creature as its source, so the
// card is drawn by the creature's controller.
//
// No simplification.
func init() {
	const grantKey = "thiefs-knife/combat-damage-draw"
	draw := WheneverThisDealsCombatDamageToAPlayer("Thief's Knife — draw a card", Do(DrawCards{N: 1}))
	Register(Spec{
		OracleID:     "4bf589ed-58f8-40ad-bcee-50cb99da02ca",
		Name:         "Thief's Knife",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{draw},
			Text:      "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(grantKey),
			AttachedIsAlso("Rogue"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Thief's Knife")},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
