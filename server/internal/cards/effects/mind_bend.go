package effects

// Mind Bend — Instant {U}:
//
//	"Change the text of target permanent by replacing all instances of
//	 one color word with another or one basic land type with another.
//	 (For example, you may change "nonblack creature" to "nongreen
//	 creature" or "forestwalk" to "islandwalk." This effect lasts
//	 indefinitely.)"
//
// ONE change, of either class: the first prompt offers the five colour
// words and the five basic land types together, and the second offers
// the other words of whichever class was picked. Permanents only, where
// Sleight of Mind and Magical Hack also reach a spell. CR 612.2,
// text_words.go, ADR 0160's 2026-10-15 amendment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8377869-f8f7-4b4b-93ad-3ad3c4826b1a",
		Name:         "Mind Bend",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		OnResolve:    changeTextOfTarget(textEitherWords, false, nothingAfterTheChange),
	})
}
