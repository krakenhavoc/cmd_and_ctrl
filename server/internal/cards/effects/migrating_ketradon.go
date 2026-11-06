package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Migrating Ketradon — Creature — Dinosaur {4}{G}{G}, 6/6 (Aetherdrift,
// #4464):
//
//	"Reach
//	 When this creature enters, you gain 4 life.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A plain enters trigger that gains life, reach as the printed keyword,
// and cycling as the hand ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d957045-acb2-4546-8c19-71511fd497ef",
		Name:            "Migrating Ketradon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{LifeGain: 4},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Migrating Ketradon — you gain 4 life", Do(GainLife{Amount: 4})),
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
