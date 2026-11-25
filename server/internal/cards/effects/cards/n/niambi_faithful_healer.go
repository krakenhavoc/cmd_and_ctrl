package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Niambi, Faithful Healer — Legendary Creature — Human Cleric {1}{W}{U}, 2/2
// (Dominaria, S82 OLD, #5530):
//
//	"When Niambi enters, you may search your library and/or graveyard for a
//	 card named Teferi, Timebender, reveal it, and put it into your hand. If
//	 you search your library this way, shuffle."
//
// The planeswalker-fetch cycle's enters trigger (MayFetchNamedFromLibraryOrGraveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c107664-a9fc-4684-8a1b-fa750401a6dd",
		Name:         "Niambi, Faithful Healer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.MayFetchNamedFromLibraryOrGraveyard("Niambi, Faithful Healer", "Teferi, Timebender")},
	})
}
