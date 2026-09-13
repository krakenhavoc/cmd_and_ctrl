package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvan Scrying — Sorcery {1}{G} (EDHREC rank 1279):
//
//	"Search your library for a land card, reveal it, put it into your
//	 hand, then shuffle."
//
// Expedition Map without the artifact: the land tutor that finds
// Cabal Coffers, Gaea's Cradle or the utility land the turn calls
// for. Type-filtered search to hand, revealed, through the S22
// chooser — b06TutorToHand with "a land card".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee24bf27-484d-4e1c-998e-6a74e3d3f6c4",
		Name:         "Sylvan Scrying",
		Completeness: CompletenessFull,
		OnResolve:    b06TutorToHand("Sylvan Scrying — a land card", func(c game.Card) bool { return c.IsLand() }),
	})
}
