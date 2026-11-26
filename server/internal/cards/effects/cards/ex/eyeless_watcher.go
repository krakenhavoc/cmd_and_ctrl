package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eyeless Watcher — Creature — Eldrazi Drone {3}{G}, 1/1 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, create two 1/1 colorless Eldrazi Scion creature
//	 tokens. They have "Sacrifice this token: Add {C}.""
//
// Two of the catalog's Eldrazi Scions, as one creation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba047b0b-d451-4566-8656-00b00611de8c",
		Name:            "Eyeless Watcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered:       []game.TriggeredAbility{effects.OldB307WhenThisEntersCreateScions("Eyeless Watcher", 2)},
	})
}
