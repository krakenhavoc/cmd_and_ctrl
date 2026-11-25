package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashiok's Forerunner — Creature — Human Wizard {3}{U}{B}, 3/3 (Theros Beyond
// Death, S82 OLD, #5514):
//
//	"Flash
//	 When this creature enters, you may search your library and/or graveyard
//	 for a card named Ashiok, Sculptor of Fears, reveal it, and put it into
//	 your hand. If you search your library this way, shuffle."
//
// Sorin's Guide's entry search (MayFetchNamedFromLibraryOrGraveyard): a "you
// may", over the library and graveyard together, the card revealed into the
// hand, the shuffle only if the library was searched. Flash rides
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "06d2a96c-4f4d-4a4a-ac48-6f0fbc9d0afd",
		Name:            "Ashiok's Forerunner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{effects.MayFetchNamedFromLibraryOrGraveyard("Ashiok's Forerunner", "Ashiok, Sculptor of Fears")},
	})
}
