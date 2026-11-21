package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	draw := effects.WheneverThisDealsCombatDamageToAPlayer("Thief's Knife — draw a card", effects.Do(effects.DrawCards{N: 1}))
	effects.Register(effects.Spec{
		OracleID:     "4bf589ed-58f8-40ad-bcee-50cb99da02ca",
		Name:         "Thief's Knife",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{draw},
			Text:      "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(grantKey),
			effects.AttachedIsAlso("Rogue"),
		},
		Triggered: []game.TriggeredAbility{effects.JobSelect("Thief's Knife")},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
