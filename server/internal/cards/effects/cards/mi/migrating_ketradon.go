package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "5d957045-acb2-4546-8c19-71511fd497ef",
		Name:            "Migrating Ketradon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{LifeGain: 4},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Migrating Ketradon — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
