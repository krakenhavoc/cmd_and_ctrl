package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territorial Scythecat — Creature — Cat {2}{G}, 2/1 (Foundations Jumpstart):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 this creature."
//
// Floral Evoker's landfall counter; trample is an engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "75d38954-6148-4d95-bf4b-cb82efcbe8df",
		Name:            "Territorial Scythecat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Territorial Scythecat — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
