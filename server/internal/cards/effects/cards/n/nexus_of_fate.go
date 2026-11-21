package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nexus of Fate — Instant {5}{U}{U}:
//
//	"Take an extra turn after this one.
//	 If Nexus of Fate would be put into a graveyard from anywhere,
//	 reveal Nexus of Fate and shuffle it into its owner's library
//	 instead."
//
// The extra turn is CR 500.7 through youTakeAnExtraTurn (Temporal
// Manipulation's body). The second sentence is Blightsteel Colossus's
// self-replacement, now the shared ShuffleIntoOwnersLibraryInstead, so
// it covers the resolution of this very spell, a counterspell, a
// discard and a mill alike, and the card is genuinely shuffled into
// its owner's library rather than left on top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c1d22d4-f28e-4041-a9b6-1575e8929b61",
		Name:         "Nexus of Fate",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.YouTakeAnExtraTurn,
		Replacements: []game.ReplacementEffect{
			effects.ShuffleIntoOwnersLibraryInstead("Nexus of Fate: shuffled into its owner's library instead"),
		},
	})
}
