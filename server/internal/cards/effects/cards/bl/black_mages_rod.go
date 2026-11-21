package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	ping := effects.WheneverYouCast(effects.Noncreature(), "Black Mage's Rod — this creature deals 1 damage to each opponent",
		effects.Do(effects.DamageEachOpponent{N: 1}))
	effects.Register(effects.Spec{
		OracleID:     "909b3d90-5762-461b-b49a-004a9e72733d",
		Name:         "Black Mage's Rod",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{ping},
			Text:      "Whenever you cast a noncreature spell, this creature deals 1 damage to each opponent.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantAbilitiesToAttached(grantKey),
			effects.AttachedIsAlso("Wizard"),
		},
		Triggered: []game.TriggeredAbility{effects.JobSelect("Black Mage's Rod")},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
