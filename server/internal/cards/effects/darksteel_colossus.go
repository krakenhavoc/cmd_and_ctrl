package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darksteel Colossus — Artifact Creature — Golem {11}, 11/11:
//
//	"Trample
//	 Indestructible
//	 If Darksteel Colossus would be put into a graveyard from anywhere,
//	 reveal Darksteel Colossus and shuffle it into its owner's library
//	 instead."
//
// Blightsteel Colossus without infect: the same replacement
// (ShuffleIntoOwnersLibraryInstead, CR 614.1a), which watches every way
// into a graveyard, a discard included, and shuffles the library once
// the card lands (ADR 0013 §5ah). It never reaches the graveyard, so
// nothing that watches a graveyard sees it. It is a static ability's
// replacement, not a CR 614.15 self-replacement, so against Rest in
// Peace's exile the card's controller, or its owner when it has none,
// chooses which applies (CR 616.1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "1b09d0cf-403c-4a15-aeee-602a1bdaf0c1",
		Name:            "Darksteel Colossus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "indestructible"},
		Replacements: []game.ReplacementEffect{
			ShuffleIntoOwnersLibraryInstead("Darksteel Colossus: shuffled into its owner's library instead"),
		},
	})
}
