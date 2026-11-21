package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farfinder — Creature — Fox {3}, 1/1 (Ikoria: Lair of Behemoths, #5512):
//
//	"Vigilance
//	 When this creature enters, you may search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle."
//
// Krosan Tusker's optional basic-land fetch with a reveal, on entry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d306de8-a93b-429b-9436-0501be97dc8d",
		Name:            "Farfinder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Farfinder — you may search your library for a basic land card, reveal it, put it into your hand, then shuffle",
				effects.IkoSearchForABasicLandToHand("Farfinder — a basic land card, to your hand")),
		},
	})
}
