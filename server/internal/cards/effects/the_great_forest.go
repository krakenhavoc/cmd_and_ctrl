package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Great Forest — Plane — Lorwyn (#4870):
//
//	"Each creature assigns combat damage equal to its toughness rather than
//	 its power.
//	 Whenever chaos ensues, creatures you control get +0/+2 and gain trample
//	 until end of turn."
//
// The static sets Characteristic.DamageByToughnessAlways on every creature
// (CR 510.1c): unlike Ghalta's "if its toughness is greater", a 5/1 assigns
// 1. Trample, deathtouch and lifelink read the toughness. The chaos boost
// and trample are applied to the creatures on the battlefield as it resolves.
//
// No simplification.
func init() {
	you := And(Creature(), YouControl())
	Register(Spec{
		OracleID:     "8c56bb41-a8d8-4b0f-bbb7-be8466dc9ed1",
		Name:         "The Great Forest",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			Label:     "Each creature assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: allCreatures,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughnessAlways = true
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("The Great Forest — creatures you control get +0/+2 and gain trample until end of turn",
				Do(
					BoostUntilEOT{Match: you, Power: 0, Toughness: 2, Label: "The Great Forest — +0/+2"},
					GrantKeywordUntilEOT{Match: you, Keywords: []string{"trample"}, Label: "The Great Forest — trample"},
				)),
		},
	})
}
