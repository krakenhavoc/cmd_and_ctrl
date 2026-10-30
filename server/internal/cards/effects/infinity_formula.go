package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infinity Formula — {2}{W} Artifact — Equipment (Marvel Super Heroes
// Commander, #4455):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+2 and has "Whenever this creature attacks,
//	 you gain 2 life."
//	 Equip {2}"
//
// The entry trigger is Falcon's Wing Harness' (AttachSourceToTarget). The
// attack trigger is a granted ability (ADR 0093), as Blue Mage's Cane's is:
// it belongs to the equipped creature, so it is that creature that attacks,
// and it goes away with the Equipment.
//
// No simplification.
func init() {
	const grantKey = "infinity-formula/gain-life"
	Register(Spec{
		OracleID:     "83669321-46cd-4805-b4ce-441c5db0f338",
		Name:         "Infinity Formula",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Infinity Formula — you gain 2 life", Do(GainLife{Amount: 2})),
			},
			Text: "Whenever this creature attacks, you gain 2 life.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
			GrantAbilitiesToAttached(grantKey),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Infinity Formula — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
