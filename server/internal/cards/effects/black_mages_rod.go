package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Mage's Rod — Artifact — Equipment {1}{B} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +1/+0, has "Whenever you cast a noncreature
//	 spell, this creature deals 1 damage to each opponent," and is a
//	 Wizard in addition to its other types.
//	 Equip {3}"
//
// The cast trigger is a granted ability (ADR 0093), as Sage's Nouliths'
// is: it belongs to the equipped creature, goes away with the Equipment,
// and has the creature as its source, so the creature deals the damage
// and "you" is the creature's controller.
//
// No simplification.
func init() {
	const grantKey = "black-mages-rod/noncreature-ping"
	ping := WheneverYouCast(Noncreature(), "Black Mage's Rod — this creature deals 1 damage to each opponent",
		Do(DamageEachOpponent{N: 1}))
	Register(Spec{
		OracleID:     "909b3d90-5762-461b-b49a-004a9e72733d",
		Name:         "Black Mage's Rod",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{ping},
			Text:      "Whenever you cast a noncreature spell, this creature deals 1 damage to each opponent.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantAbilitiesToAttached(grantKey),
			AttachedIsAlso("Wizard"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Black Mage's Rod")},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
