package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ceremonial Knife — Artifact — Equipment {1} (Innistrad: Crimson Vow):
//
//	"Equipped creature gets +1/+0 and has "Whenever this creature deals
//	 combat damage, create a Blood token."
//	 Equip {2}"
//
// All combat damage of a step is dealt at once (CR 510.2), so a creature
// that hurts a blocker and the player triggers the granted ability once
// (OncePerBatch).
func init() {
	const grantKey = "ceremonial-knife/combat-damage-blood"
	effects.Register(effects.Spec{
		OracleID:     "db221948-fcb9-4ef9-ae69-70266235dd04",
		Name:         "Ceremonial Knife",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				effects.OncePerBatch(effects.On(game.EventDealDamage, effects.S82Vow03ThisDealtCombatDamage,
					"Ceremonial Knife — create a Blood token", effects.Do(effects.S82Vow03CreateBlood(1)))),
			},
			Text: "Whenever this creature deals combat damage, create a Blood token.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantAbilitiesToAttached(grantKey),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
