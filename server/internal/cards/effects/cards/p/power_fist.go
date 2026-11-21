package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Power Fist — Artifact — Equipment {1}{G} (Fallout, #5948):
//
//	"Equipped creature has trample and "Whenever this creature deals combat
//	 damage to a player, put that many +1/+1 counters on it."
//	 Equip {2}"
//
// Biorganic Carapace's granted trigger (the ability belongs to the
// creature, so it follows the Fist's grant and not the Fist) with
// Abomination, Irradiated Brute's body: as many counters as the combat
// damage the creature dealt, put on it only while it is still the
// permanent that dealt it.
//
// No simplification.
const powerFistGrant = "power-fist/combat-damage-counters"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "bdfcf7df-8c06-4abe-af41-eccabb9fc2c8",
		Name:         "Power Fist",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: powerFistGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Power Fist — put that many +1/+1 counters on this creature",
					effects.Do(effects.Instruction(effects.AfrB18CountersEqualToDamageOnThis))),
			},
			Text: "Whenever this creature deals combat damage to a player, put that many +1/+1 counters on it.",
		}},
		Static: []game.StaticAbility{
			effects.GrantToAttached("trample"),
			effects.GrantAbilitiesToAttached(powerFistGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
