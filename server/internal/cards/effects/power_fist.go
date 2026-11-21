package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bdfcf7df-8c06-4abe-af41-eccabb9fc2c8",
		Name:         "Power Fist",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: powerFistGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Power Fist — put that many +1/+1 counters on this creature",
					Do(Instruction(afrB18CountersEqualToDamageOnThis))),
			},
			Text: "Whenever this creature deals combat damage to a player, put that many +1/+1 counters on it.",
		}},
		Static: []game.StaticAbility{
			GrantToAttached("trample"),
			GrantAbilitiesToAttached(powerFistGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
