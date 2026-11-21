package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oltec Cloud Guard — Creature — Human Soldier {3}{W}, 3/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Flying
//	 When this creature enters, create a 1/1 colorless Gnome artifact
//	 creature token."
//
// Flying is the printed keyword (CR 702.9a). The enters trigger uses the
// stack and makes the table's 1/1 colorless Gnome artifact creature,
// Tinker's Tote's token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "922a8e27-a367-40c6-84df-86ed32f71c7c",
		Name:            "Oltec Cloud Guard",
		Purpose:         game.Purpose{Tokens: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Oltec Cloud Guard — create a 1/1 colorless Gnome artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Gnome artifact"), N: 1})),
		},
	})
}
