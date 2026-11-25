package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garruk's Warsteed — Creature — Rhino {3}{G}{G}, 3/5 (Core Set 2021, #5511):
//
//	"Vigilance
//	 When this creature enters, you may search your library and/or graveyard
//	 for a card named Garruk, Savage Herald, reveal it, and put it into your
//	 hand. If you search your library this way, shuffle."
//
// The planeswalker-fetch cycle's enters trigger (library_or_graveyard_search.go):
// the zones are asked first and the shuffle follows only a library search.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d4911e00-eed1-4c8c-8779-8a3f9406ead7",
		Name:            "Garruk's Warsteed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{effects.MayFetchNamedFromLibraryOrGraveyard("Garruk's Warsteed", "Garruk, Savage Herald")},
	})
}
