package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waterwind Scout — Creature — Merfolk Scout {2}{U}, 2/2 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Flying
//	 When this creature enters, create a Map token."
//
// Spyglass Siren's text on a bigger body: a flier whose enters trigger
// makes the catalog's Map token (CR 111.10s).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "db65f874-6bf3-4fdb-a977-10b01b7ee2e9",
		Name:            "Waterwind Scout",
		Purpose:         game.Purpose{Tokens: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Waterwind Scout — create a Map token",
				effects.Do(effects.CreateToken{Template: effects.MapToken(), N: 1})),
		},
	})
}
