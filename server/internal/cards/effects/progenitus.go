package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Progenitus — Legendary Creature — Hydra Avatar
// {W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, 10/10:
//
//	"Protection from everything
//	 If Progenitus would be put into a graveyard from anywhere, reveal
//	 Progenitus and shuffle it into its owner's library instead."
//
// Protection from everything is the canonical protection token (CR
// 702.16j, ADR 0072). The second line is Blightsteel Colossus's
// replacement (ShuffleIntoOwnersLibraryInstead, CR 614.1a): from every
// zone, a discard included, shuffled in once it lands; against Rest in
// Peace its controller, or its owner when it has none, chooses (CR
// 616.1). As a commander, the move into the library is then offered
// the command zone (CR 903.9b, a replacement that can apply to the
// event the first one made, CR 616.2).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "2d7e00b6-12f0-4b03-82a6-50e3d1b5395e",
		Name:            "Progenitus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from everything"},
		Replacements: []game.ReplacementEffect{
			ShuffleIntoOwnersLibraryInstead("Progenitus: shuffled into its owner's library instead"),
		},
	})
}
