package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spyglass Siren — Creature — Siren Pirate {U}, 1/1 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Flying
//	 When this creature enters, create a Map token."
//
// Flying is the printed keyword (CR 702.9a). The enters trigger uses the
// stack and makes the catalog's Map token (CR 111.10s), as Cartographer's
// Companion does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4ec9d9eb-14e2-47ea-88a2-d2394e363784",
		Name:            "Spyglass Siren",
		Purpose:         game.Purpose{Tokens: 1},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spyglass Siren — create a Map token",
				Do(CreateToken{Template: MapToken(), N: 1})),
		},
	})
}
