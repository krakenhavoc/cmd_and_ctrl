package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relic Axe — Artifact — Equipment {2} (Foundations Jumpstart):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+1. If it's a Warrior, it gets +2/+1 instead.
//	 Equip {2}"
//
// One layer 7c modify that reads the equipped creature's subtypes as they
// stand after layer 4 (CR 613.4c): +2/+1 for a Warrior, +1/+1 otherwise.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d851312f-8b4e-4d8e-bfee-c470853be930",
		Name:         "Relic Axe",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Equipped creature gets +1/+1. If it's a Warrior, it gets +2/+1 instead.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
				if c.AllCreatureTypes {
					c.Power++
					return
				}
				for _, s := range c.Subtypes {
					if s == "Warrior" {
						c.Power++
						return
					}
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Relic Axe — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
