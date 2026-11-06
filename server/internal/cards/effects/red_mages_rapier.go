package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Red Mage's Rapier — Artifact — Equipment {1}{R} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature has "Whenever you cast a noncreature spell, this
//	 creature gets +2/+0 until end of turn" and is a Wizard in addition
//	 to its other types.
//	 Equip {3}"
//
// Black Mage's Rod's shape with a pump in place of the ping. The cast
// trigger is a granted ability (ADR 0093), so its source is the equipped
// creature: that creature gets the +2/+0, and a creature that has left
// the battlefield by the time it resolves gets nothing.
//
// No simplification.
func init() {
	const (
		grantKey = "red-mages-rapier/noncreature-pump"
		label    = "Red Mage's Rapier — this creature gets +2/+0 until end of turn"
	)
	pump := WheneverYouCast(Noncreature(), label, func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: label}.Apply(NewContext(g, item))
	})
	Register(Spec{
		OracleID:     "35ad9e15-3e00-4bd2-96cf-36a6ef49059f",
		Name:         "Red Mage's Rapier",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{pump},
			Text:      "Whenever you cast a noncreature spell, this creature gets +2/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{
			GrantAbilitiesToAttached(grantKey),
			AttachedIsAlso("Wizard"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Red Mage's Rapier")},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
