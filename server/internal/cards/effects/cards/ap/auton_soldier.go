package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const autonSoldierMyriadGrant = "auton-soldier/myriad"

// Auton Soldier — Artifact Creature — Alien Soldier {4}{U}{U}, 0/0 (Doctor
// Who, #5948):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it isn't legendary, is an artifact in addition to its
//	 other types, and has myriad."
//
// Colossal Clone Master's shape: EntersAsCopyOf over any creature with three
// exceptions that are part of the copiable values (CR 707.9b): the Legendary
// supertype is removed, the Artifact type is added, and the myriad trigger
// is granted (GrantedMyriad, as Ironwill Forger). Declining leaves the
// printed 0/0, which dies to CR 704.5f.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94357494-4ccc-43d4-b2d4-351581cbf9b6",
		Name:         "Auton Soldier",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       autonSoldierMyriadGrant,
			Text:      "Myriad",
			Triggered: []game.TriggeredAbility{effects.GrantedMyriad("Auton Soldier")},
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Auton Soldier",
				effects.AnyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.RemoveSupertype("Legendary")
					v.AddCardType("Artifact")
					v.GrantAbility(autonSoldierMyriadGrant)
				},
			),
		},
	})
}
