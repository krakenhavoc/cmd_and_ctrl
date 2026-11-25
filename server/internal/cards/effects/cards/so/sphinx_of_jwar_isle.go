package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx of Jwar Isle — Creature — Sphinx {4}{U}{U}, 5/5:
//
//	"Flying
//	 Shroud (This creature can't be the target of spells or abilities.)
//	 You may look at the top card of your library any time."
//
// Flying and shroud are printed keywords. The look is Bolas's Citadel's
// LibraryTopOwner: private to the controller, and live (the permanent is
// asked each time, so it stops when the Sphinx leaves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "3951f13e-f6dd-4d67-a088-23925baaf99b",
		Name:              "Sphinx of Jwar Isle",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"flying", "shroud"},
		LibraryTopVisible: game.LibraryTopOwner,
	})
}
