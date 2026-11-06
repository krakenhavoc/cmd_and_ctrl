package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// White Mage's Staff — Artifact — Equipment {1}{W} (Final Fantasy,
// #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/+1, has "Whenever this creature attacks,
//	 you gain 1 life," and is a Cleric in addition to its other types.
//	 Equip {3}"
//
// The attack trigger is a granted ability (ADR 0093), as Medic's
// Kitesail's is: it belongs to the equipped creature, goes away with
// the Equipment, and "you" is the creature's controller.
//
// No simplification.
func init() {
	const grantKey = "white-mages-staff/attack-life"
	Register(Spec{
		OracleID:     "af1acf07-39e2-441e-a1e0-1db535535e2e",
		Name:         "White Mage's Staff",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("White Mage's Staff — you gain 1 life", Do(GainLife{Amount: 1})),
			},
			Text: "Whenever this creature attacks, you gain 1 life.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(grantKey),
			AttachedIsAlso("Cleric"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("White Mage's Staff")},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
