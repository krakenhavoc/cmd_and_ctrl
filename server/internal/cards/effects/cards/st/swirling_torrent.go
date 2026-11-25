package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swirling Torrent — Sorcery {5}{U} (Ravnica Allegiance):
//
//	"Choose one or both —
//	 • Put target creature on top of its owner's library.
//	 • Return target creature to its owner's hand."
//
// Run Ashore's choose-one-or-both: each chosen mode names its own target, and
// the tuck is Roil Spout's (to the top of its owner's library, not a choice of
// top or bottom), shared as TuckTheModesTargetOnTopOfItsOwnersLibrary.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70f44207-338e-4031-a093-f045e8d328b2",
		Name:         "Swirling Torrent",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Put target creature on top of its owner's library.",
				effects.TargetCreature("target creature"),
				effects.TuckTheModesTargetOnTopOfItsOwnersLibrary),
			effects.ModeDoing("Return target creature to its owner's hand.",
				effects.TargetCreature("target creature"),
				effects.TyrantsScornBounce),
		),
	})
}
