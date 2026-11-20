package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrecking Beast — Creature — Beast {5}{G}{G}, 6/6 (Ravnica Allegiance,
// #5521):
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 Trample"
//
// Riot is the engine's keyword (ADR 0109 §10): the entry replacement asks
// for the counter or haste as it enters (CR 702.136a). Trample is printed
// data (CR 702.19b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "181a7b5a-3794-46ae-b6ff-48849ee1470f",
		Name:            "Wrecking Beast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot, "trample"},
	})
}
